package service

import (
	"encoding/json"
	"time"

	"distributed-scheduler-v3/internal/model"
	"distributed-scheduler-v3/internal/repository"

	"github.com/robfig/cron/v3"
)

type TaskService struct{ taskRepo *repository.TaskRepository }

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService { return &TaskService{taskRepo: taskRepo} }

type CreateDelayTaskRequest struct {
	Name           string      `json:"name" binding:"required"`
	ExecuteAt      string      `json:"execute_at" binding:"required"`
	HandlerName    string      `json:"handler_name" binding:"required"`
	Payload        interface{} `json:"payload"`
	RetryLimit     int         `json:"retry_limit"`
	TimeoutSeconds int         `json:"timeout_seconds"`
}

type CreateCronTaskRequest struct {
	Name           string      `json:"name" binding:"required"`
	CronExpr       string      `json:"cron_expr" binding:"required"`
	HandlerName    string      `json:"handler_name" binding:"required"`
	Payload        interface{} `json:"payload"`
	RetryLimit     int         `json:"retry_limit"`
	TimeoutSeconds int         `json:"timeout_seconds"`
}

func (s *TaskService) CreateDelayTask(req CreateDelayTaskRequest) (*model.Task, error) {
	execAt, err := time.ParseInLocation("2006-01-02 15:04:05", req.ExecuteAt, time.Local)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(req.Payload)
	task := &model.Task{
		Name:           req.Name,
		TaskType:       model.TaskTypeDelay,
		ExecuteAt:      &execAt,
		HandlerName:    req.HandlerName,
		Payload:        string(payload),
		RetryLimit:     req.RetryLimit,
		TimeoutSeconds: req.TimeoutSeconds,
		Status:         model.TaskStatusActive,
		NextRunTime:    &execAt,
	}
	if task.TimeoutSeconds <= 0 {
		task.TimeoutSeconds = 30
	}
	return task, s.taskRepo.Create(task)
}

func (s *TaskService) CreateCronTask(req CreateCronTaskRequest) (*model.Task, error) {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(req.CronExpr)
	if err != nil {
		return nil, err
	}
	next := schedule.Next(time.Now())
	payload, _ := json.Marshal(req.Payload)
	task := &model.Task{
		Name:           req.Name,
		TaskType:       model.TaskTypeCron,
		CronExpr:       &req.CronExpr,
		HandlerName:    req.HandlerName,
		Payload:        string(payload),
		RetryLimit:     req.RetryLimit,
		TimeoutSeconds: req.TimeoutSeconds,
		Status:         model.TaskStatusActive,
		NextRunTime:    &next,
	}
	if task.TimeoutSeconds <= 0 {
		task.TimeoutSeconds = 30
	}
	return task, s.taskRepo.Create(task)
}
