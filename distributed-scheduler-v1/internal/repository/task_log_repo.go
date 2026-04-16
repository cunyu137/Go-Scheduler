package repository

import (
	"distributed-scheduler-v1/internal/model"

	"gorm.io/gorm"
)

type TaskLogRepository struct{ db *gorm.DB }

func NewTaskLogRepository(db *gorm.DB) *TaskLogRepository { return &TaskLogRepository{db: db} }

func (r *TaskLogRepository) Create(log *model.TaskLog) error { return r.db.Create(log).Error }
