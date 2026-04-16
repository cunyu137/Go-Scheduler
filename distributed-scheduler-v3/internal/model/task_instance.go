package model

import "time"

const (
	InstanceStatusPending = "pending"
	InstanceStatusRunning = "running"
	InstanceStatusSuccess = "success"
	InstanceStatusFailed  = "failed"
	InstanceStatusTimeout = "timeout"
)

type TaskInstance struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID         int64      `gorm:"not null" json:"task_id"`
	ScheduleTime   time.Time  `gorm:"not null" json:"schedule_time"`
	Status         string     `gorm:"size:32;not null" json:"status"`
	WorkerID       *int64     `json:"worker_id,omitempty"`
	RetryCount     int        `gorm:"not null;default:0" json:"retry_count"`
	MaxRetry       int        `gorm:"not null;default:0" json:"max_retry"`
	HandlerName    string     `gorm:"size:128;not null" json:"handler_name"`
	Payload        string     `gorm:"type:json" json:"payload"`
	TimeoutSeconds int        `gorm:"not null;default:30" json:"timeout_seconds"`
	IdempotentKey  string     `gorm:"size:128;not null" json:"idempotent_key"`
	NextRetryTime  *time.Time `json:"next_retry_time,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	FinishedAt     *time.Time `json:"finished_at,omitempty"`
	ErrorMsg       string     `gorm:"type:text" json:"error_msg"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (TaskInstance) TableName() string { return "task_instances" }
