package repository

import (
	"time"

	"distributed-scheduler-v3/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct{ db *gorm.DB }

func NewTaskRepository(db *gorm.DB) *TaskRepository { return &TaskRepository{db: db} }

func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

func (r *TaskRepository) List(limit, offset int) ([]model.Task, error) {
	var out []model.Task
	if limit <= 0 {
		limit = 50
	}
	err := r.db.Order("id desc").Limit(limit).Offset(offset).Find(&out).Error
	return out, err
}

func (r *TaskRepository) FindReadyTasks(now time.Time, limit int) ([]model.Task, error) {
	var out []model.Task
	err := r.db.Where("status = ? AND next_run_time IS NOT NULL AND next_run_time <= ?", model.TaskStatusActive, now).
		Order("next_run_time asc").Limit(limit).Find(&out).Error
	return out, err
}

func (r *TaskRepository) UpdateNextRunTime(id int64, next *time.Time, status int8) error {
	return r.db.Model(&model.Task{}).Where("id = ?", id).Updates(map[string]any{
		"next_run_time": next,
		"status":        status,
	}).Error
}
