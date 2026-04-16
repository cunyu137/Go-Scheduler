package model

import "time"

const (
	WorkerStatusOnline  = "online"
	WorkerStatusOffline = "offline"
)

type Worker struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkerID        string     `gorm:"size:128;not null;uniqueIndex" json:"worker_id"`
	Address         string     `gorm:"size:255;not null" json:"address"`
	Status          string     `gorm:"size:32;not null" json:"status"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (Worker) TableName() string { return "workers" }
