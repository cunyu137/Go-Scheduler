package model

import "time"

type TaskLog struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskInstanceID int64     `gorm:"not null" json:"task_instance_id"`
	LogLevel       string    `gorm:"size:16;not null" json:"log_level"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func (TaskLog) TableName() string { return "task_logs" }
