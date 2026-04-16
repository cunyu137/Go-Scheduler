# distributed-scheduler-v3

V3 版本在 V2 的基础上新增了：

- 多 admin 下的 **Leader 选举**（基于 MySQL `GET_LOCK`）
- worker 注册与心跳
- 失联 worker 检测
- `running` 任务回收与重新派发
- admin 派发失败重试
- 仍然保留 delay / cron / retry / timeout / idempotency

## 角色

- `cmd/admin`: 调度中心。可启动多个实例，但只有 leader 会执行调度循环。
- `cmd/worker`: 执行节点。注册、发心跳、执行任务并回调结果。

## 快速运行

```bash
docker compose up -d
go mod tidy
go run ./cmd/admin
```

另开一个终端：

```bash
go run ./cmd/worker
```

如果你想模拟多 admin，只需要再开一个：

```bash
SERVER_PORT=8082 go run ./cmd/admin
```

第二个 admin 能提供 API，但只有抢到 leader lock 的实例会真正调度。

## 核心接口

### 创建 delay 任务
```bash
curl -X POST http://127.0.0.1:8081/api/v1/tasks/delay \
  -H "Content-Type: application/json" \
  -d '{
    "name":"demo print",
    "execute_at":"2026-04-16 20:30:00",
    "handler_name":"demo.print",
    "payload":{"msg":"hello v3"},
    "retry_limit":2,
    "timeout_seconds":5
  }'
```

### 创建 cron 任务
```bash
curl -X POST http://127.0.0.1:8081/api/v1/tasks/cron \
  -H "Content-Type: application/json" \
  -d '{
    "name":"demo fail once",
    "cron_expr":"*/15 * * * * *",
    "handler_name":"demo.fail_once",
    "payload":{"key":"job-1"},
    "retry_limit":2,
    "timeout_seconds":8
  }'
```

### 查看任务与实例
```bash
curl http://127.0.0.1:8081/api/v1/tasks
curl "http://127.0.0.1:8081/api/v1/task-instances?task_id=1"
curl http://127.0.0.1:8081/api/v1/workers
curl http://127.0.0.1:8081/api/v1/leader
```

## 内置 demo handlers

- `demo.print`: 正常打印并成功
- `demo.fail_once`: 第一次失败，后续成功
- `demo.timeout`: 长时间 sleep，演示 timeout

## V3 核心原理

1. admin 通过 `GET_LOCK(lock_name, 0)` 抢 leader。
2. 只有 leader 会：
   - 扫 `tasks` 生成实例
   - 扫 `task_instances` 派发到 worker
   - 扫失联 worker，回收其 `running` 任务
3. worker 周期性向 admin 发心跳。
4. admin 根据 `last_heartbeat_at + 阈值` 判断节点是否存活。
5. 如果 `running` 任务所属 worker 已失联，leader 会把任务重新置回 `pending`，等待后续重新派发。
6. worker 执行前用 Redis `SETNX` 做幂等，避免重复执行。

## 说明

这版为了本地更容易直接跑通，内部 RPC 仍使用 HTTP JSON，而不是 gRPC。
