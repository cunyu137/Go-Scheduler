package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"distributed-scheduler-v2/internal/model"
	"distributed-scheduler-v2/internal/repository"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Handler func(ctx context.Context, payload json.RawMessage) error

type Executor struct {
	handlers      map[string]Handler
	redis         *redis.Client
	instanceRepo  *repository.TaskInstanceRepository
	logRepo       *repository.TaskLogRepository
	logger        *logrus.Logger
	idempotentTTL time.Duration
}

func New(redisClient *redis.Client, instanceRepo *repository.TaskInstanceRepository, logRepo *repository.TaskLogRepository, logger *logrus.Logger, ttlSeconds int) *Executor {
	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}
	return &Executor{handlers: map[string]Handler{}, redis: redisClient, instanceRepo: instanceRepo, logRepo: logRepo, logger: logger, idempotentTTL: time.Duration(ttlSeconds) * time.Second}
}

func (e *Executor) Register(name string, handler Handler)  { e.handlers[name] = handler }
func (e *Executor) getHandler(name string) (Handler, bool) { h, ok := e.handlers[name]; return h, ok }

func (e *Executor) Execute(workerID string, instance *model.TaskInstance) (success bool, timedOut bool, message string) {
	h, ok := e.getHandler(instance.HandlerName)
	if !ok {
		return false, false, fmt.Sprintf("handler not found: %s", instance.HandlerName)
	}

	lockKey := fmt.Sprintf("task:exec:%s:attempt:%d", instance.IdempotentKey, instance.RetryCount)
	acquired, err := e.redis.SetNX(context.Background(), lockKey, workerID, e.idempotentTTL).Result()
	if err != nil {
		return false, false, fmt.Sprintf("redis setnx failed: %v", err)
	}
	if !acquired {
		return false, false, "duplicate execution skipped"
	}

	runCtx, cancel := context.WithTimeout(context.Background(), time.Duration(instance.TimeoutSeconds)*time.Second)
	defer cancel()

	resultCh := make(chan error, 1)
	go func() { resultCh <- h(runCtx, json.RawMessage(instance.Payload)) }()

	select {
	case err := <-resultCh:
		if err != nil {
			return false, false, err.Error()
		}
		return true, false, "ok"
	case <-runCtx.Done():
		return false, true, "task execution timeout"
	}
}
