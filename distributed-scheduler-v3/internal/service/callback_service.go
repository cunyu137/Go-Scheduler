package service

import (
	"fmt"
	"time"

	"distributed-scheduler-v3/internal/model"
	"distributed-scheduler-v3/internal/repository"
)

type CallbackService struct {
	instanceRepo *repository.TaskInstanceRepository
	logRepo      *repository.TaskLogRepository
}

func NewCallbackService(instanceRepo *repository.TaskInstanceRepository, logRepo *repository.TaskLogRepository) *CallbackService {
	return &CallbackService{instanceRepo: instanceRepo, logRepo: logRepo}
}

type CallbackRequest struct {
	InstanceID int64  `json:"instance_id" binding:"required"`
	Status     string `json:"status" binding:"required"`
	ErrorMsg   string `json:"error_msg"`
}

func (s *CallbackService) Handle(req CallbackRequest) error {
	inst, err := s.instanceRepo.GetByID(req.InstanceID)
	if err != nil {
		return err
	}
	switch req.Status {
	case model.InstanceStatusSuccess:
		if err := s.instanceRepo.MarkSuccess(req.InstanceID); err != nil {
			return err
		}
		return s.logRepo.Create(&model.TaskLog{TaskInstanceID: req.InstanceID, LogLevel: "INFO", Content: "worker callback success"})
	case model.InstanceStatusTimeout:
		if inst.RetryCount < inst.MaxRetry {
			nextCount := inst.RetryCount + 1
			nextRetry := time.Now().Add(time.Duration(nextCount*5) * time.Second)
			if err := s.instanceRepo.RescheduleRetry(req.InstanceID, nextCount, nextRetry, "timeout: "+req.ErrorMsg); err != nil {
				return err
			}
			return s.logRepo.Create(&model.TaskLog{TaskInstanceID: req.InstanceID, LogLevel: "WARN", Content: fmt.Sprintf("worker callback timeout, rescheduled retry=%d", nextCount)})
		}
		if err := s.instanceRepo.MarkTimeout(req.InstanceID, req.ErrorMsg); err != nil {
			return err
		}
		return s.logRepo.Create(&model.TaskLog{TaskInstanceID: req.InstanceID, LogLevel: "ERROR", Content: "worker callback timeout final"})
	case model.InstanceStatusFailed:
		if inst.RetryCount < inst.MaxRetry {
			nextCount := inst.RetryCount + 1
			nextRetry := time.Now().Add(time.Duration(nextCount*5) * time.Second)
			if err := s.instanceRepo.RescheduleRetry(req.InstanceID, nextCount, nextRetry, req.ErrorMsg); err != nil {
				return err
			}
			return s.logRepo.Create(&model.TaskLog{TaskInstanceID: req.InstanceID, LogLevel: "WARN", Content: fmt.Sprintf("worker callback failed, rescheduled retry=%d", nextCount)})
		}
		if err := s.instanceRepo.MarkFailed(req.InstanceID, req.ErrorMsg); err != nil {
			return err
		}
		return s.logRepo.Create(&model.TaskLog{TaskInstanceID: req.InstanceID, LogLevel: "ERROR", Content: "worker callback failed final"})
	default:
		return fmt.Errorf("unsupported callback status: %s", req.Status)
	}
}
