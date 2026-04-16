package model

import "time"

const (
	WorkerStatusOnline  = "online"
	WorkerStatusOffline = "offline"
)

type Worker struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Address         string    `json:"address"`
	Status          string    `json:"status"`
	LastHeartbeatAt time.Time `json:"last_heartbeat_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Worker) TableName() string { return "workers" }
