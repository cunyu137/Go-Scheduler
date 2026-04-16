package config

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type AdminConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	MySQL struct {
		DSN string `yaml:"dsn"`
	} `yaml:"mysql"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Scheduler struct {
		TaskScanIntervalSeconds      int `yaml:"task_scan_interval_seconds"`
		InstanceScanIntervalSeconds  int `yaml:"instance_scan_interval_seconds"`
		WorkerScanIntervalSeconds    int `yaml:"worker_scan_interval_seconds"`
		BatchSize                    int `yaml:"batch_size"`
		WorkerAliveSeconds           int `yaml:"worker_alive_seconds"`
		DispatchTimeoutSeconds       int `yaml:"dispatch_timeout_seconds"`
		ReclaimRunningTimeoutSeconds int `yaml:"reclaim_running_timeout_seconds"`
	} `yaml:"scheduler"`
	Leader struct {
		LockName             string `yaml:"lock_name"`
		RenewIntervalSeconds int    `yaml:"renew_interval_seconds"`
	} `yaml:"leader"`
}

type WorkerConfig struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Admin struct {
		BaseURL                  string `yaml:"base_url"`
		HeartbeatIntervalSeconds int    `yaml:"heartbeat_interval_seconds"`
		RegisterTimeoutSeconds   int    `yaml:"register_timeout_seconds"`
		CallbackTimeoutSeconds   int    `yaml:"callback_timeout_seconds"`
		WorkerID                 string `yaml:"worker_id"`
		Address                  string `yaml:"address"`
	} `yaml:"admin"`
	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`
	Executor struct {
		IdempotentKeyTTLSeconds int `yaml:"idempotent_key_ttl_seconds"`
	} `yaml:"executor"`
}

func LoadAdmin(path string) (*AdminConfig, error) {
	cfg := &AdminConfig{}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			cfg.Server.Port = p
		}
	}
	return cfg, nil
}

func LoadWorker(path string) (*WorkerConfig, error) {
	cfg := &WorkerConfig{}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
