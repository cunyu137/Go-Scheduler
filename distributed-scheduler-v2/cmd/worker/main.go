package main

import (
	"fmt"
	"os"

	"distributed-scheduler-v2/internal/config"
	"distributed-scheduler-v2/internal/executor"
	"distributed-scheduler-v2/internal/pkg/db"
	"distributed-scheduler-v2/internal/pkg/logger"
	"distributed-scheduler-v2/internal/pkg/redisutil"
	"distributed-scheduler-v2/internal/repository"
	"distributed-scheduler-v2/internal/workerapi"
	"distributed-scheduler-v2/internal/workerrunner"
)

func main() {
	cfgPath := "configs/worker.yaml"
	if val := os.Getenv("CONFIG_PATH"); val != "" {
		cfgPath = val
	}
	cfg, err := config.LoadWorker(cfgPath)
	if err != nil {
		panic(err)
	}
	log := logger.New("worker")

	database, err := db.NewMySQL(cfg.MySQL.DSN)
	if err != nil {
		panic(err)
	}
	_ = database
	rdb, err := redisutil.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		panic(err)
	}

	instanceRepo := repository.NewTaskInstanceRepository(database)
	logRepo := repository.NewTaskLogRepository(database)
	exec := executor.New(rdb, instanceRepo, logRepo, log, cfg.Redis.IdempotentKeyTTLSeconds)
	executor.RegisterBuiltinHandlers(exec)

	runner := workerrunner.New(cfg.Worker.ID, cfg.Worker.Address, cfg.Admin.BaseURL, exec, log)
	runner.StartHeartbeat(cfg.Worker.HeartbeatIntervalSeconds)

	h := workerapi.NewHandler(runner)
	r := workerapi.NewRouter(h)

	addr := fmt.Sprintf(":%d", cfg.Worker.Port)
	log.Infof("worker listening on %s", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
