# PaiBan API 使用说明

## 快速开始

### 1. 启动服务

```bash
# 方式一：一键启动
./scripts/quick-start.sh

# 方式二：手动启动
go build -o bin/paiban cmd/server/main.go
./bin/paiban
```

服务启动后访问：`http://localhost:7012`

### 2. 验证服务

```bash
# 健康检查
curl http://localhost:7012/health

# API 信息
curl http://localhost:7012/api/v1/
```

## API 端点一览

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/v1/` | GET | API 信息 |
| `/api/v1/schedule/generate` | POST | 生成排班 |
| `/api/v1/schedule/validate` | POST | 验证排班 |
| `/api/v1/constraints/templates` | GET | 约束模板 |
| `/api/v1/constraints/library` | GET | 约束库 |
| `/api/v1/stats/fairness` | POST | 公平性分析 |
| `/api/v1/stats/coverage` | POST | 覆盖率分析 |
| `/api/v1/stats/workload` | POST | 工作量统计 |
| `/api/v1/dispatch/single` | POST | 单个派单 |
| `/api/v1/dispatch/batch` | POST | 批量派单 |
| `/metrics` | GET | Prometheus 指标 |

## 核心 API 使用示例

### 1. 生成排班

```bash
curl -X POST http://localhost:7012/api/v1/schedule/generate \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "550e8400-e29b-41d4-a716-446655440000",
    "scenario": "restaurant",
    "start_date": "2024-01-15",
    "end_date": "2024-01-21",
    "employees": [
      {
        "id": "emp-001",
        "name": "张三",
        "position": "服务员",
        "skills": ["收银", "点餐"],
        "status": "active"
      },
      {
        "id": "emp-002",
        "name": "李四",
        "position": "服务员",
        "skills": ["点餐"],
        "status": "active"
      }
    ],
    "shifts": [
      {
        "id": "shift-morning",
        "name": "早班",
        "code": "M",
        "start_time": "09:00",
        "end_time": "14:00",
        "duration": 300
      },
      {
        "id": "shift-afternoon",
        "name": "午班",
        "code": "A",
        "start_time": "14:00",
        "end_time": "21:00",
        "duration": 420
      }
    ],
    "requirements": [
      {
        "shift_id": "shift-morning",
        "date": "2024-01-15",
        "min_employees": 2,
        "position": "服务员"
      },
      {
        "shift_id": "shift-afternoon",
        "date": "2024-01-15",
        "min_employees": 2,
        "position": "服务员"
      }
    ],
    "options": {
      "timeout": 30,
      "optimization_level": "balanced"
    }
  }'
```

**响应示例：**

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "schedule_id": "sch-xxx",
    "assignments": [
      {
        "employee_id": "emp-001",
        "employee_name": "张三",
        "shift_id": "shift-morning",
        "date": "2024-01-15",
        "position": "服务员"
      }
    ],
    "statistics": {
      "total_assignments": 4,
      "fill_rate": 100.0,
      "workload_variance": 0.05
    },
    "constraint_result": {
      "feasible": true,
      "hard_violations": [],
      "soft_score": 85.5
    }
  }
}
```

### 2. 验证排班

```bash
curl -X POST http://localhost:7012/api/v1/schedule/validate \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "550e8400-e29b-41d4-a716-446655440000",
    "scenario": "restaurant",
    "assignments": [
      {
        "employee_id": "emp-001",
        "shift_id": "shift-morning",
        "date": "2024-01-15"
      }
    ],
    "employees": [...],
    "shifts": [...]
  }'
```

### 3. 获取约束模板

```bash
curl http://localhost:7012/api/v1/constraints/templates
```

**响应示例：**

```json
{
  "templates": [
    {
      "scenario": "restaurant",
      "name": "餐饮门店",
      "description": "适用于餐饮行业的排班约束",
      "constraints": [
        {
          "name": "max_hours_per_day",
          "type": "hard",
          "category": "工时限制",
          "description": "每日最大工时",
          "default_value": "10"
        }
      ]
    }
  ]
}
```

### 4. 获取约束库

```bash
curl http://localhost:7012/api/v1/constraints/library
```

**响应示例：**

```json
{
  "library": [
    {
      "name": "max_hours_per_day",
      "display_name": "每日最大工时",
      "type": "hard",
      "category": "工时限制",
      "description": "限制员工每天的最大工作时长",
      "scenarios": ["restaurant", "factory", "housekeeping", "nursing"],
      "params": [
        {
          "name": "max_hours",
          "type": "int",
          "description": "最大工时(小时)",
          "default": "10",
          "min": "6",
          "max": "14"
        }
      ]
    }
  ]
}
```

### 5. 公平性分析

```bash
curl -X POST http://localhost:7012/api/v1/stats/fairness \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "550e8400-e29b-41d4-a716-446655440000",
    "start_date": "2024-01-01",
    "end_date": "2024-01-31",
    "assignments": [...]
  }'
```

### 6. 智能派单

```bash
curl -X POST http://localhost:7012/api/v1/dispatch/single \
  -H "Content-Type: application/json" \
  -d '{
    "org_id": "550e8400-e29b-41d4-a716-446655440000",
    "order": {
      "id": "order-001",
      "service_type": "cleaning",
      "address": "北京市朝阳区xxx",
      "time_window": {
        "start": "2024-01-15T09:00:00Z",
        "end": "2024-01-15T12:00:00Z"
      },
      "duration": 120,
      "required_skills": ["深度清洁"]
    },
    "available_employees": [...]
  }'
```

## 请求追踪

所有请求支持 `X-Request-ID` 头用于链路追踪：

```bash
curl -H "X-Request-ID: my-trace-123" http://localhost:7012/api/v1/schedule/generate
```

响应头中会返回相同的 `X-Request-ID`。

## 错误处理

**错误响应格式：**

```json
{
  "code": 1001,
  "message": "参数错误",
  "details": "employees 不能为空"
}
```

**常见错误码：**

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 资源不存在 |
| 2001 | 服务器内部错误 |
| 2002 | 超时 |
| 2003 | 速率限制 |

## 速率限制

API 默认限制：
- 桶容量：100 请求
- 填充速率：10 请求/秒

超限时返回 HTTP 429 状态码。

## 超时控制

排班生成支持超时设置（单位：秒）：

```json
{
  "options": {
    "timeout": 30
  }
}
```

超时后返回部分结果和未满足需求列表：

```json
{
  "data": {
    "partial": true,
    "unfilled": [
      {
        "shift_id": "shift-morning",
        "date": "2024-01-15",
        "required": 2,
        "assigned": 1,
        "shortage": 1,
        "reason": "员工不足"
      }
    ]
  }
}
```

## 场景说明

### 餐饮门店 (restaurant)

- 支持两头班
- 高峰期人员覆盖
- 健康证检查

### 工厂产线 (factory)

- 三班倒支持
- 产线24小时覆盖
- 班组协作

### 家政服务 (housekeeping)

- 服务区域匹配
- 时间窗口约束
- 路程时间优化

### 长护险 (nursing)

- 护理资质匹配
- 服务连续性
- 每日患者限制

## 前端控制台

启动可视化测试界面：

```bash
cd frontend
python3 -m http.server 8888
```

访问：http://localhost:8888

功能：
- 场景预设切换
- 请求/响应业务视图
- 约束模板在线编辑
- 约束库浏览和添加

## 更多文档

- [详细 API 指南](api-guide.md)
- [部署指南](deploy.md)
- [设计文档](design.md)
