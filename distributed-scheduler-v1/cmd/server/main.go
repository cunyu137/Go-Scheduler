package main

import (
	"fmt"
	"os"

	"distributed-scheduler-v1/internal/api"
	"distributed-scheduler-v1/internal/config"
	"distributed-scheduler-v1/internal/executor"
	"distributed-scheduler-v1/internal/pkg/db"
	"distributed-scheduler-v1/internal/pkg/logger"
	"distributed-scheduler-v1/internal/pkg/redisutil"
	"distributed-scheduler-v1/internal/repository"
	"distributed-scheduler-v1/internal/scheduler"
	"distributed-scheduler-v1/internal/service"
)

func main() {
	cfgPath := "configs/config.yaml"
	if val := os.Getenv("CONFIG_PATH"); val != "" {
		cfgPath = val
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		panic(err)
	}
	log := logger.New()

	database, err := db.NewMySQL(cfg.MySQL.DSN)
	if err != nil {
		panic(err)
	}
	rdb, err := redisutil.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		panic(err)
	}

	taskRepo := repository.NewTaskRepository(database)
	instanceRepo := repository.NewTaskInstanceRepository(database)
	logRepo := repository.NewTaskLogRepository(database)

	taskService := service.NewTaskService(taskRepo)
	exec := executor.New(rdb, instanceRepo, logRepo, log, cfg.Scheduler.IdempotentKeyTTLSeconds)
	executor.RegisterBuiltinHandlers(exec)

	sch := scheduler.New(
		taskRepo,
		instanceRepo,
		exec,
		log,
		cfg.Scheduler.TaskScanIntervalSeconds,
		cfg.Scheduler.InstanceScanIntervalSeconds,
		cfg.Scheduler.BatchSize,
	)
	sch.Start()

	h := api.NewTaskHandler(taskService, taskRepo, instanceRepo)
	r := api.NewRouter(h)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Infof("server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
