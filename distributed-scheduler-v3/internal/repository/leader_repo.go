package repository

import (
	"database/sql"

	"gorm.io/gorm"
)

type LeaderRepository struct {
	db       *gorm.DB
	lockName string
	held     bool
}

func NewLeaderRepository(db *gorm.DB, lockName string) *LeaderRepository {
	return &LeaderRepository{db: db, lockName: lockName}
}

func (r *LeaderRepository) TryAcquire() (bool, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return false, err
	}
	var got sql.NullInt64
	if err := sqlDB.QueryRow("SELECT GET_LOCK(?, 0)", r.lockName).Scan(&got); err != nil {
		return false, err
	}
	r.held = got.Valid && got.Int64 == 1
	return r.held, nil
}

func (r *LeaderRepository) Release() error {
	if !r.held {
		return nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	var out sql.NullInt64
	if err := sqlDB.QueryRow("SELECT RELEASE_LOCK(?)", r.lockName).Scan(&out); err != nil {
		return err
	}
	r.held = false
	return nil
}

func (r *LeaderRepository) IsLeader() bool { return r.held }
