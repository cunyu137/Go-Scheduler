package scheduler

import (
	"time"

	"distributed-scheduler-v3/internal/repository"

	"github.com/sirupsen/logrus"
)

type LeaderElector struct {
	repo     *repository.LeaderRepository
	logger   *logrus.Logger
	interval time.Duration
	isLeader bool
}

func NewLeaderElector(repo *repository.LeaderRepository, logger *logrus.Logger, intervalSeconds int) *LeaderElector {
	if intervalSeconds <= 0 {
		intervalSeconds = 2
	}
	return &LeaderElector{repo: repo, logger: logger, interval: time.Duration(intervalSeconds) * time.Second}
}

func (l *LeaderElector) Start() {
	go func() {
		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()
		for {
			got, err := l.repo.TryAcquire()
			if err != nil {
				l.logger.WithError(err).Warn("leader acquire failed")
				l.isLeader = false
			} else {
				l.isLeader = got
			}
			<-ticker.C
		}
	}()
}

func (l *LeaderElector) IsLeader() bool { return l.isLeader }
