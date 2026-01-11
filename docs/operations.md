# 排班引擎运维指南

## 1. 系统架构

### 1.1 组件概览

```
┌─────────────────────────────────────────────────────────────┐
│                     Load Balancer (Nginx)                    │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  Paiban API   │     │  Paiban API   │     │  Paiban API   │
│   (Pod 1)     │     │   (Pod 2)     │     │   (Pod N)     │
└───────────────┘     └───────────────┘     └───────────────┘
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  PostgreSQL   │     │     Redis     │     │  Prometheus   │
│   (主库)       │     │   (缓存)      │     │   (监控)       │
└───────────────┘     └───────────────┘     └───────────────┘
```

### 1.2 端口配置

| 服务 | 端口 | 说明 |
|------|------|------|
| Paiban API | 7012 | 主服务端口 |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| Prometheus | 9090 | 监控 |

## 2. 部署指南

### 2.1 Docker 部署

```bash
# 构建镜像
docker build -f deployments/Dockerfile -t paiban:latest .

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f paiban
```

### 2.2 Kubernetes 部署

```bash
# 应用配置
kubectl apply -f deployments/k8s/

# 检查状态
kubectl get pods -l app=paiban

# 查看日志
kubectl logs -f deployment/paiban
```

### 2.3 健康检查

```bash
# 健康检查端点
curl http://localhost:7012/health

# 版本信息
curl http://localhost:7012/version

# Prometheus 指标
curl http://localhost:7012/metrics
```

## 3. 配置管理

### 3.1 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| APP_PORT | 7012 | 服务端口 |
| APP_ENV | development | 运行环境 |
| DB_HOST | localhost | 数据库主机 |
| DB_PORT | 5432 | 数据库端口 |
| DB_NAME | paiban | 数据库名称 |
| DB_USER | paiban | 数据库用户 |
| DB_PASSWORD | - | 数据库密码 |
| REDIS_HOST | localhost | Redis主机 |
| REDIS_PORT | 6379 | Redis端口 |
| LOG_LEVEL | info | 日志级别 |

### 3.2 配置文件

配置文件位于 `configs/app.yaml`:

```yaml
app:
  name: paiban
  port: 7012
  env: production

database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  name: ${DB_NAME}
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  max_connections: 100
  idle_connections: 10

redis:
  host: ${REDIS_HOST}
  port: ${REDIS_PORT}
  db: 0

logging:
  level: info
  format: json
```

## 4. 监控告警

### 4.1 关键指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| http_requests_total | HTTP请求总数 | - |
| http_request_duration_seconds | 请求延迟 | p99 > 5s |
| schedule_generations_total | 排班生成次数 | - |
| schedule_generation_duration_seconds | 排班生成耗时 | > 30s |
| constraint_evaluations_total | 约束评估次数 | - |
| go_goroutines | Goroutine数量 | > 10000 |
| go_memstats_alloc_bytes | 内存使用 | > 2GB |

### 4.2 Prometheus 查询示例

```promql
# 请求QPS
rate(http_requests_total[5m])

# 请求延迟P99
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))

# 排班生成成功率
sum(rate(schedule_generations_total{status="success"}[5m])) / 
sum(rate(schedule_generations_total[5m]))

# 内存使用趋势
go_memstats_alloc_bytes
```

### 4.3 告警规则

```yaml
groups:
  - name: paiban-alerts
    rules:
      - alert: HighRequestLatency
        expr: histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "请求延迟过高"
          
      - alert: ScheduleGenerationFailed
        expr: increase(schedule_generations_total{status="error"}[5m]) > 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "排班生成失败率过高"
          
      - alert: HighMemoryUsage
        expr: go_memstats_alloc_bytes > 2e9
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "内存使用过高"
```

## 5. 日志管理

### 5.1 日志格式

```json
{
  "level": "info",
  "time": "2026-01-11T10:30:00Z",
  "caller": "handler/schedule.go:45",
  "message": "排班生成完成",
  "org_id": "xxx",
  "employees": 50,
  "shifts": 100,
  "duration_ms": 1234
}
```

### 5.2 日志级别

| 级别 | 说明 |
|------|------|
| debug | 调试信息 |
| info | 一般信息 |
| warn | 警告信息 |
| error | 错误信息 |
| fatal | 致命错误 |

### 5.3 日志轮转

使用 logrotate 配置:

```
/var/log/paiban/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 paiban paiban
}
```

## 6. 数据库维护

### 6.1 备份策略

```bash
# 每日全量备份
pg_dump -h localhost -U paiban -d paiban > backup_$(date +%Y%m%d).sql

# 增量备份（使用WAL归档）
archive_mode = on
archive_command = 'cp %p /backup/wal/%f'
```

### 6.2 常用SQL

```sql
-- 查看活动连接
SELECT * FROM pg_stat_activity WHERE datname = 'paiban';

-- 查看表大小
SELECT relname, pg_size_pretty(pg_relation_size(relid))
FROM pg_stat_user_tables
ORDER BY pg_relation_size(relid) DESC;

-- 清理过期数据
DELETE FROM assignments WHERE created_at < NOW() - INTERVAL '1 year';
DELETE FROM service_orders WHERE created_at < NOW() - INTERVAL '6 months';

-- 重建索引
REINDEX TABLE assignments;
```

### 6.3 性能优化

```sql
-- 分析表统计信息
ANALYZE assignments;
ANALYZE service_orders;

-- 检查慢查询
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

## 7. 故障排查

### 7.1 常见问题

#### 服务无法启动

1. 检查端口占用: `lsof -i :7012`
2. 检查数据库连接: `pg_isready -h localhost -p 5432`
3. 检查配置文件: `cat configs/app.yaml`

#### 排班生成超时

1. 检查约束复杂度
2. 减少员工/班次数量
3. 调整优化参数

#### 内存占用过高

1. 检查Goroutine泄漏: `/debug/pprof/goroutine`
2. 检查大对象分配: `/debug/pprof/heap`
3. 调整GC参数: `GOGC=100`

### 7.2 调试工具

```bash
# CPU分析
go tool pprof http://localhost:7012/debug/pprof/profile?seconds=30

# 内存分析
go tool pprof http://localhost:7012/debug/pprof/heap

# Goroutine分析
go tool pprof http://localhost:7012/debug/pprof/goroutine

# 跟踪
go tool trace trace.out
```

## 8. 扩容指南

### 8.1 水平扩展

```yaml
# Kubernetes HPA配置
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: paiban-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: paiban
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

### 8.2 资源配置建议

| 场景 | CPU | 内存 | 实例数 |
|------|-----|------|--------|
| 开发 | 0.5 | 512MB | 1 |
| 测试 | 1 | 1GB | 2 |
| 生产(小) | 2 | 2GB | 3 |
| 生产(大) | 4 | 4GB | 5+ |

## 9. 安全配置

### 9.1 API密钥管理

```bash
# 生成API密钥
curl -X POST http://localhost:7012/api/v1/admin/keys \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -d '{"name": "客户A", "scopes": ["schedule", "dispatch"]}'

# 撤销API密钥
curl -X DELETE http://localhost:7012/api/v1/admin/keys/{key_id} \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### 9.2 网络安全

- 仅允许内网访问管理端点
- 使用HTTPS加密传输
- 配置防火墙规则
- 启用请求频率限制

## 10. 版本更新

### 10.1 滚动更新

```bash
# Kubernetes滚动更新
kubectl set image deployment/paiban paiban=paiban:v2.0.0

# 查看更新状态
kubectl rollout status deployment/paiban

# 回滚
kubectl rollout undo deployment/paiban
```

### 10.2 数据库迁移

```bash
# 执行迁移
make migrate-up

# 回滚迁移
make migrate-down

# 查看迁移状态
make migrate-status
```

