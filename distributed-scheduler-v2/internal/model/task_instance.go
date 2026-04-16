package model

import "time"

const (
	InstanceStatusPending    = "pending"
	InstanceStatusDispatched = "dispatched"
	InstanceStatusRunning    = "running"
	InstanceStatusSuccess    = "success"
	InstanceStatusFailed     = "failed"
	InstanceStatusTimeout    = "timeout"
)

type TaskInstance struct {
	ID             int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID         int64      `json:"task_id"`
	ScheduleTime   time.Time  `json:"schedule_time"`
	Status         string     `json:"status"`
	RetryCount     int        `json:"retry_count"`
	MaxRetry       int        `json:"max_retry"`
	NextRetryTime  *time.Time `json:"next_retry_time"`
	HandlerName    string     `json:"handler_name"`
	Payload        string     `json:"payload"`
	TimeoutSeconds int        `json:"timeout_seconds"`
	IdempotentKey  string     `json:"idempotent_key"`
	WorkerID       *string    `json:"worker_id"`
	StartedAt      *time.Time `json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	ErrorMsg       *string    `json:"error_msg"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (TaskInstance) TableName() string { return "task_instances" }
