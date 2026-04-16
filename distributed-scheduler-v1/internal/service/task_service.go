package service

import (
	"fmt"
	"time"

	"distributed-scheduler-v1/internal/model"
	"distributed-scheduler-v1/internal/repository"

	"github.com/robfig/cron/v3"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

type CreateDelayTaskRequest struct {
	Name           string    `json:"name" binding:"required"`
	ExecuteAt      string    `json:"execute_at" binding:"required"`
	HandlerName    string    `json:"handler_name" binding:"required"`
	Payload        string    `json:"payload"`
	RetryLimit     int       `json:"retry_limit"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	_              time.Time `json:"-"`
}

type CreateCronTaskRequest struct {
	Name           string `json:"name" binding:"required"`
	CronExpr       string `json:"cron_expr" binding:"required"`
	HandlerName    string `json:"handler_name" binding:"required"`
	Payload        string `json:"payload"`
	RetryLimit     int    `json:"retry_limit"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

func (s *TaskService) CreateDelayTask(req CreateDelayTaskRequest) (*model.Task, error) {
	execAt, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExecuteAt, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid execute_at format, use 2006-01-02 15:04:05")
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 30
	}
	task := &model.Task{
		Name:           req.Name,
		TaskType:       model.TaskTypeDelay,
		ExecuteAt:      &execAt,
		HandlerName:    req.HandlerName,
		Payload:        normalizePayload(req.Payload),
		RetryLimit:     req.RetryLimit,
		TimeoutSeconds: req.TimeoutSeconds,
		Status:         model.TaskStatusActive,
		NextRunTime:    &execAt,
	}
	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) CreateCronTask(req CreateCronTaskRequest) (*model.Task, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(req.CronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}
	next := schedule.Next(time.Now())
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 30
	}
	cronExpr := req.CronExpr
	task := &model.Task{
		Name:           req.Name,
		TaskType:       model.TaskTypeCron,
		CronExpr:       &cronExpr,
		HandlerName:    req.HandlerName,
		Payload:        normalizePayload(req.Payload),
		RetryLimit:     req.RetryLimit,
		TimeoutSeconds: req.TimeoutSeconds,
		Status:         model.TaskStatusActive,
		NextRunTime:    &next,
	}
	if err := s.taskRepo.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func normalizePayload(payload string) string {
	if payload == "" {
		return `{}`
	}
	return payload
}
