package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"distributed-scheduler-v3/internal/config"
	"distributed-scheduler-v3/internal/executor"
	"distributed-scheduler-v3/internal/pkg/db"
	"distributed-scheduler-v3/internal/pkg/logger"
	"distributed-scheduler-v3/internal/pkg/redisutil"
	"distributed-scheduler-v3/internal/repository"
	"distributed-scheduler-v3/internal/workerapi"
	"distributed-scheduler-v3/internal/workerrunner"
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

	rdb := redisutil.New(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	var logRepo *repository.TaskLogRepository
	if adminDSN := os.Getenv("MYSQL_DSN"); adminDSN != "" {
		if database, err := db.NewMySQL(adminDSN); err == nil {
			logRepo = repository.NewTaskLogRepository(database)
		}
	}

	exec := executor.New(rdb, logRepo, log, cfg.Executor.IdempotentKeyTTLSeconds)
	executor.RegisterBuiltinHandlers(exec)

	runner := workerrunner.New(exec, cfg.Admin.BaseURL, cfg.Admin.CallbackTimeoutSeconds)
	h := workerapi.NewHandler(runner)
	r := workerapi.NewRouter(h)

	go registerAndHeartbeat(cfg.Admin.BaseURL, cfg.Admin.WorkerID, cfg.Admin.Address, cfg.Admin.RegisterTimeoutSeconds, cfg.Admin.HeartbeatIntervalSeconds, log)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Infof("worker listening on %s", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}

func registerAndHeartbeat(baseURL, workerID, address string, registerTimeoutSec, hbIntervalSec int, log interface{ Infof(string, ...any); Warnf(string, ...any) }) {
	baseURL = strings.TrimRight(baseURL, "/")
	client := &http.Client{Timeout: time.Duration(registerTimeoutSec) * time.Second}
	body, _ := json.Marshal(map[string]string{"worker_id": workerID, "address": address})
	_, _ = client.Post(baseURL+"/internal/workers/register", "application/json", bytes.NewReader(body))
	if hbIntervalSec <= 0 {
		hbIntervalSec = 5
	}
	ticker := time.NewTicker(time.Duration(hbIntervalSec) * time.Second)
	defer ticker.Stop()
	for {
		body, _ := json.Marshal(map[string]string{"worker_id": workerID, "address": address})
		_, _ = client.Post(baseURL+"/internal/workers/heartbeat", "application/json", bytes.NewReader(body))
		<-ticker.C
	}
}
