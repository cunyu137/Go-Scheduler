package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"distributed-scheduler-v3/internal/model"
	"distributed-scheduler-v3/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Handler func(ctx context.Context, payload json.RawMessage) error

type CallbackFunc func(status, errMsg string)

type Executor struct {
	handlers      map[string]Handler
	redis         *redis.Client
	logRepo       *repository.TaskLogRepository
	logger        *logrus.Logger
	idempotentTTL time.Duration
}

func New(rdb *redis.Client, logRepo *repository.TaskLogRepository, logger *logrus.Logger, ttlSeconds int) *Executor {
	if ttlSeconds <= 0 {
		ttlSeconds = 3600
	}
	return &Executor{
		handlers:      map[string]Handler{},
		redis:         rdb,
		logRepo:       logRepo,
		logger:        logger,
		idempotentTTL: time.Duration(ttlSeconds) * time.Second,
	}
}

func (e *Executor) Register(name string, h Handler) { e.handlers[name] = h }

func (e *Executor) Execute(instanceID int64, handlerName string, payload string, timeoutSeconds int, idempotentKey string, callback CallbackFunc) {
	h, ok := e.handlers[handlerName]
	if !ok {
		callback(model.InstanceStatusFailed, "handler not found: "+handlerName)
		return
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}
	lockKey := "task:exec:" + idempotentKey
	acquired, err := e.redis.SetNX(context.Background(), lockKey, "1", e.idempotentTTL).Result()
	if err != nil {
		callback(model.InstanceStatusFailed, "idempotency redis error: "+err.Error())
		return
	}
	if !acquired {
		callback(model.InstanceStatusSuccess, "duplicate dispatch skipped by idempotency key")
		return
	}

	runCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	_ = e.appendLog(instanceID, "INFO", "worker task started")
	resultCh := make(chan error, 1)
	go func() {
		resultCh <- h(runCtx, json.RawMessage(payload))
	}()

	select {
	case err := <-resultCh:
		if err != nil {
			_ = e.appendLog(instanceID, "ERROR", err.Error())
			callback(model.InstanceStatusFailed, err.Error())
			return
		}
		_ = e.appendLog(instanceID, "INFO", "worker task succeeded")
		callback(model.InstanceStatusSuccess, "")
	case <-runCtx.Done():
		msg := fmt.Sprintf("task execution timeout after %d seconds", timeoutSeconds)
		_ = e.appendLog(instanceID, "ERROR", msg)
		callback(model.InstanceStatusTimeout, msg)
	}
}

func (e *Executor) appendLog(instanceID int64, level, content string) error {
	if e.logRepo == nil {
		return nil
	}
	return e.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, LogLevel: level, Content: content})
}
