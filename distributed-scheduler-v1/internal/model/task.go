package model

import "time"

const (
	TaskTypeDelay = "delay"
	TaskTypeCron  = "cron"

	TaskStatusActive = 1
	TaskStatusPaused = 2
)

type Task struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string     `gorm:"size:128;not null" json:"name"`
	TaskType       string     `gorm:"size:32;not null" json:"task_type"`
	CronExpr       *string    `gorm:"size:64" json:"cron_expr,omitempty"`
	ExecuteAt      *time.Time `json:"execute_at,omitempty"`
	HandlerName    string     `gorm:"size:128;not null" json:"handler_name"`
	Payload        string     `gorm:"type:json" json:"payload"`
	RetryLimit     int        `gorm:"not null;default:0" json:"retry_limit"`
	TimeoutSeconds int        `gorm:"not null;default:30" json:"timeout_seconds"`
	Status         int        `gorm:"not null;default:1" json:"status"`
	NextRunTime    *time.Time `json:"next_run_time,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Task) TableName() string {
	return "tasks"
}
