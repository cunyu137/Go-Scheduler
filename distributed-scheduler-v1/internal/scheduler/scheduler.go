package scheduler

import (
	"fmt"
	"time"

	"distributed-scheduler-v1/internal/executor"
	"distributed-scheduler-v1/internal/model"
	"distributed-scheduler-v1/internal/repository"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	taskRepo     *repository.TaskRepository
	instanceRepo *repository.TaskInstanceRepository
	executor     *executor.Executor
	logger       *logrus.Logger
	taskInterval time.Duration
	instInterval time.Duration
	batchSize    int
}

func New(taskRepo *repository.TaskRepository, instanceRepo *repository.TaskInstanceRepository, exec *executor.Executor, logger *logrus.Logger, taskIntervalSec, instIntervalSec, batchSize int) *Scheduler {
	return &Scheduler{
		taskRepo:     taskRepo,
		instanceRepo: instanceRepo,
		executor:     exec,
		logger:       logger,
		taskInterval: time.Duration(taskIntervalSec) * time.Second,
		instInterval: time.Duration(instIntervalSec) * time.Second,
		batchSize:    batchSize,
	}
}

func (s *Scheduler) Start() {
	go s.loopGenerateInstances()
	go s.loopExecuteInstances()
}

func (s *Scheduler) loopGenerateInstances() {
	ticker := time.NewTicker(s.taskInterval)
	defer ticker.Stop()
	for {
		s.generateInstances()
		<-ticker.C
	}
}

func (s *Scheduler) loopExecuteInstances() {
	ticker := time.NewTicker(s.instInterval)
	defer ticker.Stop()
	for {
		s.executeInstances()
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
		inst := &model.TaskInstance{
			TaskID:         task.ID,
			ScheduleTime:   scheduleTime.Local(),
			Status:         model.InstanceStatusPending,
			RetryCount:     0,
			MaxRetry:       task.RetryLimit,
			HandlerName:    task.HandlerName,
			Payload:        task.Payload,
			TimeoutSeconds: task.TimeoutSeconds,
			IdempotentKey:  fmt.Sprintf("task:%d:%s", task.ID, scheduleTime.Format("20060102150405")),
		}
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

func (s *Scheduler) executeInstances() {
	now := time.Now()
	instances, err := s.instanceRepo.FindRunnable(now, s.batchSize)
	if err != nil {
		s.logger.WithError(err).Error("find runnable instances failed")
		return
	}
	for _, inst := range instances {
		ok, err := s.instanceRepo.MarkRunning(inst.ID)
		if err != nil {
			s.logger.WithError(err).WithField("instance_id", inst.ID).Error("mark running failed")
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
		go s.executor.Execute(fresh)
	}
}
