package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	MySQL     MySQLConfig     `yaml:"mysql"`
	Redis     RedisConfig     `yaml:"redis"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type MySQLConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type SchedulerConfig struct {
	TaskScanIntervalSeconds     int `yaml:"task_scan_interval_seconds"`
	InstanceScanIntervalSeconds int `yaml:"instance_scan_interval_seconds"`
	BatchSize                   int `yaml:"batch_size"`
	IdempotentKeyTTLSeconds     int `yaml:"idempotent_key_ttl_seconds"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Scheduler.TaskScanIntervalSeconds == 0 {
		cfg.Scheduler.TaskScanIntervalSeconds = 1
	}
	if cfg.Scheduler.InstanceScanIntervalSeconds == 0 {
		cfg.Scheduler.InstanceScanIntervalSeconds = 1
	}
	if cfg.Scheduler.BatchSize == 0 {
		cfg.Scheduler.BatchSize = 50
	}
	if cfg.Scheduler.IdempotentKeyTTLSeconds == 0 {
		cfg.Scheduler.IdempotentKeyTTLSeconds = 3600
	}
	return &cfg, nil
}
