CREATE DATABASE IF NOT EXISTS scheduler DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE scheduler;

CREATE TABLE IF NOT EXISTS tasks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(128) NOT NULL,
    task_type VARCHAR(32) NOT NULL COMMENT 'delay / cron',
    cron_expr VARCHAR(64) DEFAULT NULL,
    execute_at DATETIME DEFAULT NULL,
    handler_name VARCHAR(128) NOT NULL,
    payload JSON DEFAULT NULL,
    retry_limit INT NOT NULL DEFAULT 0,
    timeout_seconds INT NOT NULL DEFAULT 30,
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=active, 2=paused, 3=deleted',
    next_run_time DATETIME DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_next_run_time (next_run_time),
    INDEX idx_status (status)
);

CREATE TABLE IF NOT EXISTS task_instances (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id BIGINT NOT NULL,
    schedule_time DATETIME NOT NULL,
    status VARCHAR(32) NOT NULL COMMENT 'pending/running/success/failed/timeout',
    retry_count INT NOT NULL DEFAULT 0,
    max_retry INT NOT NULL DEFAULT 0,
    next_retry_time DATETIME DEFAULT NULL,
    handler_name VARCHAR(128) NOT NULL,
    payload JSON DEFAULT NULL,
    timeout_seconds INT NOT NULL DEFAULT 30,
    idempotent_key VARCHAR(128) NOT NULL,
    started_at DATETIME DEFAULT NULL,
    finished_at DATETIME DEFAULT NULL,
    error_msg TEXT DEFAULT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_task_schedule (task_id, schedule_time),
    UNIQUE KEY uk_idempotent_key (idempotent_key),
    INDEX idx_status_schedule (status, schedule_time),
    INDEX idx_status_retry (status, next_retry_time),
    INDEX idx_task_id (task_id)
);

CREATE TABLE IF NOT EXISTS task_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_instance_id BIGINT NOT NULL,
    log_level VARCHAR(16) NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_instance_id (task_instance_id)
);
