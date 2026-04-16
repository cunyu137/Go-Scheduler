package repository

import (
	"time"

	"distributed-scheduler-v3/internal/model"
	"gorm.io/gorm"
)

type WorkerRepository struct{ db *gorm.DB }

func NewWorkerRepository(db *gorm.DB) *WorkerRepository { return &WorkerRepository{db: db} }

func (r *WorkerRepository) Upsert(workerID, address string) error {
	now := time.Now()
	var existing model.Worker
	err := r.db.Where("worker_id = ?", workerID).First(&existing).Error
	if err == nil {
		return r.db.Model(&existing).Updates(map[string]any{
			"address":            address,
			"status":             model.WorkerStatusOnline,
			"last_heartbeat_at":  &now,
		}).Error
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	w := &model.Worker{WorkerID: workerID, Address: address, Status: model.WorkerStatusOnline, LastHeartbeatAt: &now}
	return r.db.Create(w).Error
}

func (r *WorkerRepository) Heartbeat(workerID, address string) error {
	now := time.Now()
	return r.db.Model(&model.Worker{}).Where("worker_id = ?", workerID).Updates(map[string]any{
		"address":            address,
		"status":             model.WorkerStatusOnline,
		"last_heartbeat_at":  &now,
	}).Error
}

func (r *WorkerRepository) FindAlive(cutoff time.Time, limit int) ([]model.Worker, error) {
	var out []model.Worker
	if limit <= 0 {
		limit = 100
	}
	err := r.db.Where("status = ? AND last_heartbeat_at IS NOT NULL AND last_heartbeat_at >= ?", model.WorkerStatusOnline, cutoff).
		Order("id asc").Limit(limit).Find(&out).Error
	return out, err
}

func (r *WorkerRepository) MarkOfflineBefore(cutoff time.Time) error {
	return r.db.Model(&model.Worker{}).
		Where("status = ? AND (last_heartbeat_at IS NULL OR last_heartbeat_at < ?)", model.WorkerStatusOnline, cutoff).
		Update("status", model.WorkerStatusOffline).Error
}

func (r *WorkerRepository) FindOfflineSince(cutoff time.Time, limit int) ([]model.Worker, error) {
	var out []model.Worker
	if limit <= 0 {
		limit = 100
	}
	err := r.db.Where("(last_heartbeat_at IS NULL OR last_heartbeat_at < ?)", cutoff).
		Order("id asc").Limit(limit).Find(&out).Error
	return out, err
}

func (r *WorkerRepository) List(limit, offset int) ([]model.Worker, error) {
	var out []model.Worker
	if limit <= 0 {
		limit = 100
	}
	err := r.db.Order("id asc").Limit(limit).Offset(offset).Find(&out).Error
	return out, err
}
