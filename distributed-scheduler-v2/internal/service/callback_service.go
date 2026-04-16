package service

import (
	"fmt"
	"time"

	"distributed-scheduler-v2/internal/model"
	"distributed-scheduler-v2/internal/repository"
)

type CallbackService struct {
	instanceRepo *repository.TaskInstanceRepository
	logRepo      *repository.TaskLogRepository
}

func NewCallbackService(instanceRepo *repository.TaskInstanceRepository, logRepo *repository.TaskLogRepository) *CallbackService {
	return &CallbackService{instanceRepo: instanceRepo, logRepo: logRepo}
}

func (s *CallbackService) MarkStarted(instanceID int64, workerID string) error {
	if err := s.instanceRepo.MarkRunningByID(instanceID); err != nil {
		return err
	}
	return s.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, WorkerID: &workerID, LogLevel: "INFO", Content: "worker started task"})
}

func (s *CallbackService) Finish(instanceID int64, workerID string, success bool, timedOut bool, message string, retryCount int, maxRetry int) error {
	if success {
		if err := s.instanceRepo.MarkSuccess(instanceID); err != nil {
			return err
		}
		return s.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, WorkerID: &workerID, LogLevel: "INFO", Content: "worker finished task successfully"})
	}
	if timedOut {
		if retryCount < maxRetry {
			nextRetry := time.Now().Add(time.Duration(retryCount+1) * 5 * time.Second)
			if err := s.instanceRepo.RescheduleRetry(instanceID, retryCount+1, nextRetry, fmt.Sprintf("timeout: %s", message)); err != nil {
				return err
			}
			return s.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, WorkerID: &workerID, LogLevel: "WARN", Content: fmt.Sprintf("task timeout, rescheduled to %s", nextRetry.Format(time.RFC3339))})
		}
		if err := s.instanceRepo.MarkTimeout(instanceID, message); err != nil {
			return err
		}
		return s.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, WorkerID: &workerID, LogLevel: "ERROR", Content: "task timeout and no retries left"})
	}
	if retryCount < maxRetry {
		nextRetry := time.Now().Add(time.Duration(retryCount+1) * 5 * time.Second)
		if err := s.instanceRepo.RescheduleRetry(instanceID, retryCount+1, nextRetry, message); err != nil {
			return err
		}
		return s.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, WorkerID: &workerID, LogLevel: "WARN", Content: fmt.Sprintf("task failed, rescheduled to %s", nextRetry.Format(time.RFC3339))})
	}
	if err := s.instanceRepo.MarkFailed(instanceID, message); err != nil {
		return err
	}
	return s.logRepo.Create(&model.TaskLog{TaskInstanceID: instanceID, WorkerID: &workerID, LogLevel: "ERROR", Content: "task failed and no retries left"})
}
