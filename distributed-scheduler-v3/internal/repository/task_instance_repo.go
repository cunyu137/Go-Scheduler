package repository

import (
	"time"

	"distributed-scheduler-v3/internal/model"
	"gorm.io/gorm"
)

type TaskInstanceRepository struct{ db *gorm.DB }

func NewTaskInstanceRepository(db *gorm.DB) *TaskInstanceRepository { return &TaskInstanceRepository{db: db} }

func (r *TaskInstanceRepository) Create(inst *model.TaskInstance) error {
	return r.db.Create(inst).Error
}

func (r *TaskInstanceRepository) GetByID(id int64) (*model.TaskInstance, error) {
	var inst model.TaskInstance
	err := r.db.First(&inst, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &inst, nil
}

func (r *TaskInstanceRepository) ListByTaskID(taskID int64, limit, offset int) ([]model.TaskInstance, error) {
	var out []model.TaskInstance
	q := r.db.Order("id desc")
	if taskID > 0 {
		q = q.Where("task_id = ?", taskID)
	}
	if limit <= 0 {
		limit = 50
	}
	err := q.Limit(limit).Offset(offset).Find(&out).Error
	return out, err
}

func (r *TaskInstanceRepository) FindRunnable(now time.Time, limit int) ([]model.TaskInstance, error) {
	var out []model.TaskInstance
	err := r.db.Where("status = ? AND (next_retry_time IS NULL OR next_retry_time <= ?)", model.InstanceStatusPending, now).
		Order("schedule_time asc").Limit(limit).Find(&out).Error
	return out, err
}

func (r *TaskInstanceRepository) MarkRunningWithWorker(id int64, workerDBID int64) (bool, error) {
	now := time.Now()
	tx := r.db.Model(&model.TaskInstance{}).
		Where("id = ? AND status = ?", id, model.InstanceStatusPending).
		Updates(map[string]any{
			"status":     model.InstanceStatusRunning,
			"worker_id":  workerDBID,
			"started_at": &now,
		})
	return tx.RowsAffected > 0, tx.Error
}

func (r *TaskInstanceRepository) MarkSuccess(id int64) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.InstanceStatusSuccess,
		"finished_at": &now,
		"error_msg":   "",
	}).Error
}

func (r *TaskInstanceRepository) MarkTimeout(id int64, msg string) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.InstanceStatusTimeout,
		"finished_at": &now,
		"error_msg":   msg,
	}).Error
}

func (r *TaskInstanceRepository) MarkFailed(id int64, msg string) error {
	now := time.Now()
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":      model.InstanceStatusFailed,
		"finished_at": &now,
		"error_msg":   msg,
	}).Error
}

func (r *TaskInstanceRepository) RescheduleRetry(id int64, retryCount int, nextRetry time.Time, msg string) error {
	return r.db.Model(&model.TaskInstance{}).Where("id = ?", id).Updates(map[string]any{
		"status":          model.InstanceStatusPending,
		"retry_count":     retryCount,
		"next_retry_time": &nextRetry,
		"error_msg":       msg,
		"worker_id":       nil,
	}).Error
}

func (r *TaskInstanceRepository) RequeueRunningByWorker(workerDBID int64, staleBefore time.Time, msg string) (int64, error) {
	tx := r.db.Model(&model.TaskInstance{}).
		Where("status = ? AND worker_id = ? AND updated_at <= ?", model.InstanceStatusRunning, workerDBID, staleBefore).
		Updates(map[string]any{
			"status":          model.InstanceStatusPending,
			"worker_id":       nil,
			"next_retry_time": time.Now(),
			"error_msg":       msg,
		})
	return tx.RowsAffected, tx.Error
}

func (r *TaskInstanceRepository) RequeueRunningByDeadWorkers(workerIDs []int64, staleBefore time.Time, msg string) (int64, error) {
	if len(workerIDs) == 0 {
		return 0, nil
	}
	tx := r.db.Model(&model.TaskInstance{}).
		Where("status = ? AND worker_id IN ? AND updated_at <= ?", model.InstanceStatusRunning, workerIDs, staleBefore).
		Updates(map[string]any{
			"status":          model.InstanceStatusPending,
			"worker_id":       nil,
			"next_retry_time": time.Now(),
			"error_msg":       msg,
		})
	return tx.RowsAffected, tx.Error
}
