package repository

import (
	"errors"
	"strings"
	"time"

	"distributed-scheduler-v1/internal/model"

	"gorm.io/gorm"
)

type TaskInstanceRepository struct{ db *gorm.DB }

func NewTaskInstanceRepository(db *gorm.DB) *TaskInstanceRepository {
	return &TaskInstanceRepository{db: db}
}

func (r *TaskInstanceRepository) Create(instance *model.TaskInstance) error {
	return r.db.Create(instance).Error
}

func (r *TaskInstanceRepository) ListByTaskID(taskID int64, page, pageSize int) ([]model.TaskInstance, int64, error) {
	var instances []model.TaskInstance
	var total int64
	q := r.db.Model(&model.TaskInstance{}).Where("task_id = ?", taskID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&instances).Error; err != nil {
		return nil, 0, err
	}
	return instances, total, nil
}

func (r *TaskInstanceRepository) FindRunnable(now time.Time, limit int) ([]model.TaskInstance, error) {
	var instances []model.TaskInstance
	err := r.db.Where("status = ? AND (next_retry_time IS NULL OR next_retry_time <= ?)", model.InstanceStatusPending, now).
		Order("schedule_time asc").Limit(limit).Find(&instances).Error
	return instances, err
}

func (r *TaskInstanceRepository) MarkRunning(id int64) (bool, error) {
	now := time.Now()
	res := r.db.Model(&model.TaskInstance{}).
		Where("id = ? AND status = ?", id, model.InstanceStatusPending).
		Updates(map[string]any{"status": model.InstanceStatusRunning, "started_at": now, "updated_at": now})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *TaskInstanceRepository) MarkSuccess(id int64) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.InstanceStatusSuccess,
		"finished_at": now,
		"error_msg":   "",
		"updated_at":  now,
	}).Error
}

func (r *TaskInstanceRepository) MarkTimeout(id int64, msg string) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.InstanceStatusTimeout,
		"finished_at": now,
		"error_msg":   msg,
		"updated_at":  now,
	}).Error
}

func (r *TaskInstanceRepository) RescheduleRetry(id int64, retryCount int, nextRetry time.Time, msg string) error {
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":          model.InstanceStatusPending,
		"retry_count":     retryCount,
		"next_retry_time": nextRetry,
		"error_msg":       msg,
		"updated_at":      time.Now(),
	}).Error
}

func (r *TaskInstanceRepository) MarkFailed(id int64, msg string) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.InstanceStatusFailed,
		"finished_at": now,
		"error_msg":   msg,
		"updated_at":  now,
	}).Error
}

func (r *TaskInstanceRepository) ExistsByTaskAndSchedule(taskID int64, schedule time.Time) (bool, error) {
	var count int64
	err := r.db.Model(&model.TaskInstance{}).Where("task_id = ? AND schedule_time = ?", taskID, schedule).Count(&count).Error
	return count > 0, err
}

func (r *TaskInstanceRepository) GetByID(id int64) (*model.TaskInstance, error) {
	var inst model.TaskInstance
	if err := r.db.First(&inst, id).Error; err != nil {
		return nil, err
	}
	return &inst, nil
}

func IsDuplicateErr(err error) bool {
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) || containsDuplicate(err.Error()))
}

func containsDuplicate(s string) bool {
	return s != "" && (strings.Contains(s, "Duplicate entry") || strings.Contains(s, "duplicated key"))
}
