package model

import "time"

type TaskLog struct {
	ID             int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskInstanceID int64     `json:"task_instance_id"`
	WorkerID       *string   `json:"worker_id"`
	LogLevel       string    `json:"log_level"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func (TaskLog) TableName() string { return "task_logs" }
