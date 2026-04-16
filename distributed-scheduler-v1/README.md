# distributed-scheduler-v1

一个可直接运行的 Go 任务调度系统 V1：
- 支持延迟任务
- 支持 Cron 定时任务
- 支持失败重试
- 支持任务超时控制
- 支持 Redis 幂等控制
- 支持任务实例查询

## 运行方式

### 1. 启动 MySQL 和 Redis
```bash
docker compose up -d
```

### 2. 启动服务
```bash
go mod tidy
go run ./cmd/server
```

### 3. 健康检查
```bash
curl http://127.0.0.1:8080/health
```

## 创建任务示例

### 延迟任务 demo.print
```bash
curl -X POST http://127.0.0.1:8080/api/v1/tasks/delay \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"print once",
    "execute_at":"2026-04-16 23:59:00",
    "handler_name":"demo.print",
    "payload":"{\"msg\":\"hello scheduler\"}",
    "retry_limit":0,
    "timeout_seconds":5
  }'
```

### Cron 任务 demo.fail_once
```bash
curl -X POST http://127.0.0.1:8080/api/v1/tasks/cron \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"fail once cron",
    "cron_expr":"*/15 * * * * *",
    "handler_name":"demo.fail_once",
    "payload":"{\"key\":\"job-1\"}",
    "retry_limit":2,
    "timeout_seconds":5
  }'
```

### 超时任务 demo.timeout
```bash
curl -X POST http://127.0.0.1:8080/api/v1/tasks/delay \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"timeout test",
    "execute_at":"2026-04-16 23:59:10",
    "handler_name":"demo.timeout",
    "payload":"{}",
    "retry_limit":0,
    "timeout_seconds":3
  }'
```

## 查询任务
```bash
curl 'http://127.0.0.1:8080/api/v1/tasks?page=1&page_size=10'
```

## 查询任务实例
```bash
curl 'http://127.0.0.1:8080/api/v1/task-instances?task_id=1&page=1&page_size=10'
```

## 内置 handler
- `demo.print`: 成功执行
- `demo.fail_once`: 第一次失败，重试后成功
- `demo.timeout`: 故意超时
