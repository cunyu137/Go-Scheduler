package repository

import (
	"time"

	"distributed-scheduler-v2/internal/model"

	"gorm.io/gorm"
)

type TaskInstanceRepository struct{ db *gorm.DB }

func NewTaskInstanceRepository(db *gorm.DB) *TaskInstanceRepository {
	return &TaskInstanceRepository{db: db}
}

func (r *TaskInstanceRepository) Create(inst *model.TaskInstance) error {
	return r.db.Create(inst).Error
}

func (r *TaskInstanceRepository) ListByTaskID(taskID int64, page, pageSize int) ([]model.TaskInstance, int64, error) {
	var items []model.TaskInstance
	var total int64
	q := r.db.Model(&model.TaskInstance{}).Where("task_id = ?", taskID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *TaskInstanceRepository) FindRunnable(now time.Time, limit int) ([]model.TaskInstance, error) {
	var items []model.TaskInstance
	err := r.db.Where("status = ? AND (next_retry_time IS NULL OR next_retry_time <= ?)", model.InstanceStatusPending, now).
		Order("schedule_time asc").Limit(limit).Find(&items).Error
	return items, err
}

func (r *TaskInstanceRepository) FindDispatchedTimeout(before time.Time, limit int) ([]model.TaskInstance, error) {
	var items []model.TaskInstance
	err := r.db.Where("status = ? AND updated_at <= ?", model.InstanceStatusDispatched, before).
		Order("updated_at asc").Limit(limit).Find(&items).Error
	return items, err
}

func (r *TaskInstanceRepository) MarkDispatched(instanceID int64, workerID string) (bool, error) {
	tx := r.db.Model(&model.TaskInstance{}).Where("id = ? AND status = ?", instanceID, model.InstanceStatusPending).
		Updates(map[string]any{"status": model.InstanceStatusDispatched, "worker_id": workerID, "updated_at": time.Now()})
	return tx.RowsAffected > 0, tx.Error
}

func (r *TaskInstanceRepository) MarkRunningByID(instanceID int64) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", instanceID).
		Updates(map[string]any{"status": model.InstanceStatusRunning, "started_at": now, "updated_at": now}).Error
}

func (r *TaskInstanceRepository) MarkSuccess(instanceID int64) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", instanceID).
		Updates(map[string]any{"status": model.InstanceStatusSuccess, "finished_at": now, "updated_at": now, "error_msg": nil}).Error
}

func (r *TaskInstanceRepository) MarkTimeout(instanceID int64, msg string) error {
	now := time.Now()
	e := msg
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", instanceID).
		Updates(map[string]any{"status": model.InstanceStatusTimeout, "finished_at": now, "updated_at": now, "error_msg": &e}).Error
}

func (r *TaskInstanceRepository) MarkFailed(instanceID int64, msg string) error {
	now := time.Now()
	e := msg
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", instanceID).
		Updates(map[string]any{"status": model.InstanceStatusFailed, "finished_at": now, "updated_at": now, "error_msg": &e}).Error
}

func (r *TaskInstanceRepository) RescheduleRetry(instanceID int64, retryCount int, nextRetry time.Time, msg string) error {
	e := msg
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", instanceID).
		Updates(map[string]any{"status": model.InstanceStatusPending, "retry_count": retryCount, "next_retry_time": nextRetry, "updated_at": time.Now(), "error_msg": &e}).Error
}

func (r *TaskInstanceRepository) GetByID(id int64) (*model.TaskInstance, error) {
	var item model.TaskInstance
	if err := r.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}
