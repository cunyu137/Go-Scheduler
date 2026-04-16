# distributed-scheduler-v2

V2 版本将 V1 的“调度”和“执行”拆成两个独立服务：
- `admin`：任务管理、任务扫描、实例生成、实例派发、结果回写
- `worker`：注册、心跳、接收任务、执行 handler

## 这版和 V1 的区别
- V1: 一个进程同时负责 API + 调度 + 执行
- V2: admin 只负责调度，worker 只负责执行
- 新增 `workers` 表、worker 注册与心跳、远程派发、结果回传

## 运行步骤

### 1. 启动 MySQL 和 Redis
```bash
docker compose up -d
```

### 2. 启动 admin
```bash
go mod tidy
go run ./cmd/admin
```

### 3. 启动 worker
另开一个终端：
```bash
go run ./cmd/worker
```

### 4. 健康检查
```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8090/health
```

## 创建任务示例

### 延迟任务
```bash
curl -X POST http://127.0.0.1:8080/api/v1/tasks/delay   -H 'Content-Type: application/json'   -d '{
    "name":"print once",
    "execute_at":"2026-04-16 23:59:00",
    "handler_name":"demo.print",
    "payload":"{"msg":"hello from v2"}",
    "retry_limit":0,
    "timeout_seconds":5
  }'
```

### Cron 任务（第一次失败，之后成功）
```bash
curl -X POST http://127.0.0.1:8080/api/v1/tasks/cron   -H 'Content-Type: application/json'   -d '{
    "name":"fail once cron",
    "cron_expr":"*/15 * * * * *",
    "handler_name":"demo.fail_once",
    "payload":"{"key":"job-1"}",
    "retry_limit":2,
    "timeout_seconds":5
  }'
```

### 查询 worker
```bash
curl http://127.0.0.1:8080/api/v1/workers
```

### 查询任务实例
```bash
curl 'http://127.0.0.1:8080/api/v1/task-instances?task_id=1&page=1&page_size=20'
```

## 任务主流程
1. 用户创建任务 -> 写入 `tasks`
2. admin 扫描 `tasks` -> 生成 `task_instances(pending)`
3. admin 扫描 `task_instances` -> 选择一个在线 worker
4. admin 将实例标记为 `dispatched`
5. admin 通过 HTTP 调 worker `/internal/execute`
6. worker 收到任务后异步执行并调用 admin 回调：
   - `/internal/instances/:id/start`
   - `/internal/instances/:id/finish`
7. admin 根据结果把实例更新为 `running/success/failed/timeout`，必要时重试

## 内置 handler
- `demo.print`: 正常成功
- `demo.fail_once`: 第一次失败，后续成功
- `demo.timeout`: 配合 timeout 测试超时

## 说明
为了让你更容易本地直接跑通，这版 V2 内部服务间通信使用了 HTTP JSON，而不是 gRPC。整体的“调度中心 + worker + 注册心跳 + 远程派发”架构已经是 V2 版本，后续要升到 gRPC 只需要替换内部 RPC 层即可。
