package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AdminConfig struct {
	Server    ServerConfig    `yaml:"server"`
	MySQL     MySQLConfig     `yaml:"mysql"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
}

type WorkerConfigFile struct {
	Worker WorkerConfig `yaml:"worker"`
	Admin  AdminTarget  `yaml:"admin"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}
type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}
type AdminTarget struct {
	BaseURL string `yaml:"base_url"`
}
type WorkerConfig struct {
	ID                       string `yaml:"id"`
	Address                  string `yaml:"address"`
	Port                     int    `yaml:"port"`
	HeartbeatIntervalSeconds int    `yaml:"heartbeat_interval_seconds"`
}
type RedisConfig struct {
	Addr                    string `yaml:"addr"`
	Password                string `yaml:"password"`
	DB                      int    `yaml:"db"`
	IdempotentKeyTTLSeconds int    `yaml:"idempotent_key_ttl_seconds"`
}
type SchedulerConfig struct {
	TaskScanIntervalSeconds     int `yaml:"task_scan_interval_seconds"`
	InstanceScanIntervalSeconds int `yaml:"instance_scan_interval_seconds"`
	BatchSize                   int `yaml:"batch_size"`
	WorkerAliveSeconds          int `yaml:"worker_alive_seconds"`
	DispatchTimeoutSeconds      int `yaml:"dispatch_timeout_seconds"`
}

func LoadAdmin(path string) (*AdminConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AdminConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server.Port <= 0 {
		return nil, fmt.Errorf("invalid server.port")
	}
	return &cfg, nil
}

func LoadWorker(path string) (*WorkerConfigFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg WorkerConfigFile
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	if cfg.Worker.ID == "" || cfg.Worker.Address == "" || cfg.Worker.Port <= 0 || cfg.Admin.BaseURL == "" {
		return nil, fmt.Errorf("invalid worker config")
	}
	return &cfg, nil
}
