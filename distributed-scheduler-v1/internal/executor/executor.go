package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"distributed-scheduler-v1/internal/model"
	"distributed-scheduler-v1/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Handler func(ctx context.Context, payload json.RawMessage) error

type Executor struct {
	mu            sync.RWMutex
	handlers      map[string]Handler
	redis         *redis.Client
	instanceRepo  *repository.TaskInstanceRepository
	logRepo       *repository.TaskLogRepository
	logger        *logrus.Logger
	idempotentTTL time.Duration
}

func New(rdb *redis.Client, instanceRepo *repository.TaskInstanceRepository, logRepo *repository.TaskLogRepository, logger *logrus.Logger, ttlSeconds int) *Executor {
	return &Executor{
		handlers:      make(map[string]Handler),
		redis:         rdb,
		instanceRepo:  instanceRepo,
		logRepo:       logRepo,
		logger:        logger,
		idempotentTTL: time.Duration(ttlSeconds) * time.Second,
	}
}

func (e *Executor) Register(name string, h Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[name] = h
}

func (e *Executor) Execute(instance *model.TaskInstance) {
	h, ok := e.getHandler(instance.HandlerName)
	if !ok {
		e.fail(instance, fmt.Sprintf("handler not found: %s", instance.HandlerName))
		return
	}
	ctx := context.Background()
	lockKey := "task:exec:" + instance.IdempotentKey
	acquired, err := e.redis.SetNX(ctx, lockKey, "1", e.idempotentTTL).Result()
	if err != nil {
		e.fail(instance, fmt.Sprintf("acquire idempotent lock failed: %v", err))
		return
	}
	if !acquired {
		e.logger.WithField("instance_id", instance.ID).Warn("skip duplicated execution due to idempotent key")
		return
	}

	runCtx, cancel := context.WithTimeout(context.Background(), time.Duration(instance.TimeoutSeconds)*time.Second)
	defer cancel()

	_ = e.appendLog(instance.ID, "INFO", "task started")
	resultCh := make(chan error, 1)
	go func() {
		resultCh <- h(runCtx, json.RawMessage(instance.Payload))
	}()

	select {
	case err := <-resultCh:
		if err != nil {
			e.handleFailure(instance, err)
			return
		}
		if err := e.instanceRepo.MarkSuccess(instance.ID); err != nil {
			e.logger.WithError(err).WithField("instance_id", instance.ID).Error("mark success failed")
		}
		_ = e.appendLog(instance.ID, "INFO", "task finished successfully")
		e.logger.WithField("instance_id", instance.ID).Info("task executed successfully")
	case <-runCtx.Done():
		msg := "task execution timeout"
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			msg = runCtx.Err().Error()
		}
		if err := e.instanceRepo.MarkTimeout(instance.ID, msg); err != nil {
			e.logger.WithError(err).WithField("instance_id", instance.ID).Error("mark timeout failed")
		}
		_ = e.appendLog(instance.ID, "ERROR", msg)
		e.logger.WithField("instance_id", instance.ID).Warn(msg)
	}
}

func (e *Executor) handleFailure(instance *model.TaskInstance, err error) {
	msg := err.Error()
	_ = e.appendLog(instance.ID, "ERROR", msg)
	if instance.RetryCount < instance.MaxRetry {
		nextCount := instance.RetryCount + 1
		nextRetry := time.Now().Add(time.Duration(nextCount*5) * time.Second)
		if repoErr := e.instanceRepo.RescheduleRetry(instance.ID, nextCount, nextRetry, msg); repoErr != nil {
			e.logger.WithError(repoErr).WithField("instance_id", instance.ID).Error("reschedule retry failed")
			return
		}
		e.logger.WithFields(logrus.Fields{"instance_id": instance.ID, "retry_count": nextCount, "next_retry": nextRetry}).Warn("task failed and will retry")
		return
	}
	e.fail(instance, msg)
}

func (e *Executor) fail(instance *model.TaskInstance, msg string) {
	if err := e.instanceRepo.MarkFailed(instance.ID, msg); err != nil {
		e.logger.WithError(err).WithField("instance_id", instance.ID).Error("mark failed failed")
	}
	_ = e.appendLog(instance.ID, "ERROR", msg)
	e.logger.WithField("instance_id", instance.ID).Error(msg)
}

func (e *Executor) appendLog(taskInstanceID int64, level, content string) error {
	return e.logRepo.Create(&model.TaskLog{TaskInstanceID: taskInstanceID, LogLevel: level, Content: content})
}

func (e *Executor) getHandler(name string) (Handler, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	h, ok := e.handlers[name]
	return h, ok
}
