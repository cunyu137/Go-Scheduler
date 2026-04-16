package scheduler

import (
	"fmt"
	"sync/atomic"
	"time"

	"distributed-scheduler-v2/internal/adminclient"
	"distributed-scheduler-v2/internal/model"
	"distributed-scheduler-v2/internal/repository"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	taskRepo           *repository.TaskRepository
	instanceRepo       *repository.TaskInstanceRepository
	workerRepo         *repository.WorkerRepository
	dispatcher         *adminclient.WorkerDispatcher
	logger             *logrus.Logger
	taskInterval       time.Duration
	instInterval       time.Duration
	batchSize          int
	workerAliveSeconds int
	rr                 uint64
}

func New(taskRepo *repository.TaskRepository, instanceRepo *repository.TaskInstanceRepository, workerRepo *repository.WorkerRepository, dispatcher *adminclient.WorkerDispatcher, logger *logrus.Logger, taskIntervalSec, instIntervalSec, batchSize, workerAliveSeconds int) *Scheduler {
	return &Scheduler{taskRepo: taskRepo, instanceRepo: instanceRepo, workerRepo: workerRepo, dispatcher: dispatcher, logger: logger, taskInterval: time.Duration(taskIntervalSec) * time.Second, instInterval: time.Duration(instIntervalSec) * time.Second, batchSize: batchSize, workerAliveSeconds: workerAliveSeconds}
}

func (s *Scheduler) Start() {
	go s.loopGenerateInstances()
	go s.loopDispatchInstances()
}

func (s *Scheduler) loopGenerateInstances() {
	ticker := time.NewTicker(s.taskInterval)
	defer ticker.Stop()
	for {
		s.generateInstances()
		<-ticker.C
	}
}

func (s *Scheduler) loopDispatchInstances() {
	ticker := time.NewTicker(s.instInterval)
	defer ticker.Stop()
	for {
		s.dispatchInstances()
		<-ticker.C
	}
}

func (s *Scheduler) generateInstances() {
	now := time.Now()
	tasks, err := s.taskRepo.FindReadyTasks(now, s.batchSize)
	if err != nil {
		s.logger.WithError(err).Error("find ready tasks failed")
		return
	}
	for _, task := range tasks {
		scheduleTime := task.NextRunTime
		if scheduleTime == nil {
			continue
		}
		inst := &model.TaskInstance{TaskID: task.ID, ScheduleTime: scheduleTime.Local(), Status: model.InstanceStatusPending, RetryCount: 0, MaxRetry: task.RetryLimit, HandlerName: task.HandlerName, Payload: task.Payload, TimeoutSeconds: task.TimeoutSeconds, IdempotentKey: fmt.Sprintf("task:%d:%s", task.ID, scheduleTime.Format("20060102150405"))}
		if err := s.instanceRepo.Create(inst); err != nil && !repository.IsDuplicateErr(err) {
			s.logger.WithError(err).WithField("task_id", task.ID).Error("create task instance failed")
			continue
		}
		if task.TaskType == model.TaskTypeDelay {
			if err := s.taskRepo.UpdateNextRunTime(task.ID, nil, model.TaskStatusPaused); err != nil {
				s.logger.WithError(err).WithField("task_id", task.ID).Error("pause delay task failed")
			}
			continue
		}
		if task.TaskType == model.TaskTypeCron && task.CronExpr != nil {
			parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
			schedule, err := parser.Parse(*task.CronExpr)
			if err != nil {
				s.logger.WithError(err).WithField("task_id", task.ID).Error("parse cron failed")
				continue
			}
			next := schedule.Next(*scheduleTime)
			if err := s.taskRepo.UpdateNextRunTime(task.ID, &next, model.TaskStatusActive); err != nil {
				s.logger.WithError(err).WithField("task_id", task.ID).Error("update cron next run failed")
			}
		}
	}
}

func (s *Scheduler) dispatchInstances() {
	cutoff := time.Now().Add(-time.Duration(s.workerAliveSeconds) * time.Second)
	_ = s.workerRepo.MarkOfflineBefore(cutoff)
	workers, err := s.workerRepo.FindAlive(cutoff, 100)
	if err != nil {
		s.logger.WithError(err).Error("find workers failed")
		return
	}
	if len(workers) == 0 {
		return
	}

	items, err := s.instanceRepo.FindRunnable(time.Now(), s.batchSize)
	if err != nil {
		s.logger.WithError(err).Error("find runnable instances failed")
		return
	}
	for _, inst := range items {
		w := workers[int(atomic.AddUint64(&s.rr, 1))%len(workers)]
		ok, err := s.instanceRepo.MarkDispatched(inst.ID, w.ID)
		if err != nil {
			s.logger.WithError(err).WithField("instance_id", inst.ID).Error("mark dispatched failed")
			continue
		}
		if !ok {
			continue
		}
		fresh, err := s.instanceRepo.GetByID(inst.ID)
		if err != nil {
			s.logger.WithError(err).WithField("instance_id", inst.ID).Error("reload instance failed")
			continue
		}
		if err := s.dispatcher.Dispatch(w, *fresh); err != nil {
			s.logger.WithError(err).WithField("instance_id", inst.ID).WithField("worker_id", w.ID).Warn("dispatch failed, reschedule")
			nextRetry := time.Now().Add(3 * time.Second)
			_ = s.instanceRepo.RescheduleRetry(inst.ID, fresh.RetryCount, nextRetry, "dispatch failed: "+err.Error())
		}
	}
}
