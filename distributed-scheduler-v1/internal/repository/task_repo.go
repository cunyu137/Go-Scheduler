package repository

import (
	"time"

	"distributed-scheduler-v1/internal/model"

	"gorm.io/gorm"
)

type TaskRepository struct{ db *gorm.DB }

func NewTaskRepository(db *gorm.DB) *TaskRepository { return &TaskRepository{db: db} }

func (r *TaskRepository) Create(task *model.Task) error { return r.db.Create(task).Error }

func (r *TaskRepository) List(page, pageSize int, taskType string, status int) ([]model.Task, int64, error) {
	var tasks []model.Task
	var total int64
	q := r.db.Model(&model.Task{})
	if taskType != "" {
		q = q.Where("task_type = ?", taskType)
	}
	if status != 0 {
		q = q.Where("status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}
	return tasks, total, nil
}

func (r *TaskRepository) FindReadyTasks(now time.Time, limit int) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("status = ? AND next_run_time IS NOT NULL AND next_run_time <= ?", model.TaskStatusActive, now).
		Order("next_run_time asc").Limit(limit).Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) UpdateNextRunTime(taskID int64, next *time.Time, status int) error {
	updates := map[string]any{"status": status, "updated_at": time.Now()}
	if next != nil {
		updates["next_run_time"] = *next
	} else {
		updates["next_run_time"] = nil
	}
	return r.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

func (r *TaskRepository) GetByID(id int64) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}
