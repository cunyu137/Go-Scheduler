package main

import (
	"fmt"
	"os"

	"distributed-scheduler-v3/internal/adminapi"
	"distributed-scheduler-v3/internal/adminclient"
	"distributed-scheduler-v3/internal/config"
	"distributed-scheduler-v3/internal/pkg/db"
	"distributed-scheduler-v3/internal/pkg/logger"
	"distributed-scheduler-v3/internal/repository"
	"distributed-scheduler-v3/internal/scheduler"
	"distributed-scheduler-v3/internal/service"
)

func main() {
	cfgPath := "configs/admin.yaml"
	if val := os.Getenv("CONFIG_PATH"); val != "" {
		cfgPath = val
	}
	cfg, err := config.LoadAdmin(cfgPath)
	if err != nil {
		panic(err)
	}
	log := logger.New("admin")

	database, err := db.NewMySQL(cfg.MySQL.DSN)
	if err != nil {
		panic(err)
	}

	taskRepo := repository.NewTaskRepository(database)
	instanceRepo := repository.NewTaskInstanceRepository(database)
	logRepo := repository.NewTaskLogRepository(database)
	workerRepo := repository.NewWorkerRepository(database)
	leaderRepo := repository.NewLeaderRepository(database, cfg.Leader.LockName)

	taskService := service.NewTaskService(taskRepo)
	callbackService := service.NewCallbackService(instanceRepo, logRepo)
	dispatcher := adminclient.NewWorkerDispatcher(cfg.Scheduler.DispatchTimeoutSeconds)
	elector := scheduler.NewLeaderElector(leaderRepo, log, cfg.Leader.RenewIntervalSeconds)
	elector.Start()

	sch := scheduler.New(
		taskRepo,
		instanceRepo,
		workerRepo,
		dispatcher,
		log,
		cfg.Scheduler.TaskScanIntervalSeconds,
		cfg.Scheduler.InstanceScanIntervalSeconds,
		cfg.Scheduler.WorkerScanIntervalSeconds,
		cfg.Scheduler.BatchSize,
		cfg.Scheduler.WorkerAliveSeconds,
		cfg.Scheduler.ReclaimRunningTimeoutSeconds,
		elector.IsLeader,
	)
	sch.Start()

	h := adminapi.NewHandler(taskService, taskRepo, instanceRepo, workerRepo, callbackService, elector, log)
	r := adminapi.NewRouter(h)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Infof("admin listening on %s", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
