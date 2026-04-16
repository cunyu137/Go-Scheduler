package repository

import (
	"time"

	"distributed-scheduler-v2/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WorkerRepository struct{ db *gorm.DB }

func NewWorkerRepository(db *gorm.DB) *WorkerRepository { return &WorkerRepository{db: db} }

func (r *WorkerRepository) UpsertHeartbeat(id, address string) error {
	now := time.Now()
	w := model.Worker{ID: id, Address: address, Status: model.WorkerStatusOnline, LastHeartbeatAt: now}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{"address": address, "status": model.WorkerStatusOnline, "last_heartbeat_at": now, "updated_at": now}),
	}).Create(&w).Error
}

func (r *WorkerRepository) List(page, pageSize int) ([]model.Worker, int64, error) {
	var items []model.Worker
	var total int64
	q := r.db.Model(&model.Worker{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id asc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *WorkerRepository) FindAlive(aliveAfter time.Time, limit int) ([]model.Worker, error) {
	var items []model.Worker
	err := r.db.Where("status = ? AND last_heartbeat_at >= ?", model.WorkerStatusOnline, aliveAfter).
		Order("last_heartbeat_at desc").Limit(limit).Find(&items).Error
	return items, err
}

func (r *WorkerRepository) MarkOfflineBefore(cutoff time.Time) error {
	return r.db.Model(&model.Worker{}).Where("last_heartbeat_at < ?", cutoff).
		Updates(map[string]any{"status": model.WorkerStatusOffline, "updated_at": time.Now()}).Error
}
