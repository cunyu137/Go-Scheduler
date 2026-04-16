package model

import "time"

const (
	TaskTypeDelay = "delay"
	TaskTypeCron  = "cron"
)

const (
	TaskStatusActive  = 1
	TaskStatusPaused  = 2
	TaskStatusDeleted = 3
)

type Task struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string     `json:"name"`
	TaskType       string     `json:"task_type"`
	CronExpr       *string    `json:"cron_expr"`
	ExecuteAt      *time.Time `json:"execute_at"`
	HandlerName    string     `json:"handler_name"`
	Payload        string     `json:"payload"`
	RetryLimit     int        `json:"retry_limit"`
	TimeoutSeconds int        `json:"timeout_seconds"`
	Status         int        `json:"status"`
	NextRunTime    *time.Time `json:"next_run_time"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }
