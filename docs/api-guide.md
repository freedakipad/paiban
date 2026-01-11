# 排班引擎API接入指南

## 1. 概述

排班引擎提供RESTful API，支持四大业务场景的智能排班和派单服务：
- 餐饮行业排班
- 工厂流水线排班
- 家政服务派单
- 长护险护理服务

## 2. 快速开始

### 2.1 获取API密钥

联系系统管理员获取API密钥，密钥格式如下：
```
pk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 2.2 认证方式

在请求头中添加API密钥：

```http
Authorization: Bearer pk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

或使用 X-API-Key 头：

```http
X-API-Key: pk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

### 2.3 基础URL

```
生产环境: https://api.paiban.com/api/v1
测试环境: https://test-api.paiban.com/api/v1
本地开发: http://localhost:7012/api/v1
```

## 3. 排班API

### 3.1 生成排班

**请求**

```http
POST /api/v1/schedule/generate
Content-Type: application/json
```

**请求体**

```json
{
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "scenario": "restaurant",
  "start_date": "2026-01-11",
  "end_date": "2026-01-17",
  "employees": [
    {
      "id": "emp-001",
      "name": "张三",
      "skills": ["cooking", "service"],
      "certifications": ["health_cert"],
      "preferences": {
        "preferred_shifts": ["morning", "afternoon"],
        "unavailable_dates": ["2026-01-15"]
      }
    }
  ],
  "shifts": [
    {
      "id": "shift-001",
      "name": "早班",
      "type": "morning",
      "start_time": "08:00",
      "end_time": "14:00",
      "required_count": 3,
      "required_skills": ["service"]
    }
  ],
  "constraints": [
    {
      "type": "MaxHoursPerDay",
      "params": {"max_hours": 10},
      "weight": 100
    },
    {
      "type": "MinRestBetweenShifts",
      "params": {"min_hours": 8},
      "weight": 100
    }
  ],
  "options": {
    "optimize": true,
    "max_iterations": 1000,
    "timeout_seconds": 60
  }
}
```

**响应**

```json
{
  "success": true,
  "schedule_id": "sched-2026011101",
  "assignments": [
    {
      "employee_id": "emp-001",
      "employee_name": "张三",
      "shift_id": "shift-001",
      "date": "2026-01-11",
      "start_time": "08:00",
      "end_time": "14:00"
    }
  ],
  "statistics": {
    "total_shifts": 21,
    "assigned_shifts": 20,
    "coverage_rate": 95.2,
    "constraint_violations": 0
  },
  "warnings": []
}
```

### 3.2 验证排班

**请求**

```http
POST /api/v1/schedule/validate
Content-Type: application/json
```

**请求体**

```json
{
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "assignments": [...],
  "constraints": [...]
}
```

**响应**

```json
{
  "valid": true,
  "violations": [],
  "score": 98.5
}
```

### 3.3 获取约束模板

**请求**

```http
GET /api/v1/constraints/templates?scenario=restaurant
```

**响应**

```json
{
  "scenario": "restaurant",
  "constraints": [
    {
      "type": "MaxHoursPerDay",
      "category": "hard",
      "default_params": {"max_hours": 10},
      "default_weight": 100,
      "description": "每日最大工时限制"
    },
    {
      "type": "PeakHoursMinStaff",
      "category": "hard",
      "default_params": {
        "peak_hours": {"11:00-13:00": 5, "18:00-20:00": 6}
      },
      "default_weight": 100,
      "description": "高峰时段最低人力"
    }
  ]
}
```

## 4. 派单API

### 4.1 单订单派单

**请求**

```http
POST /api/v1/dispatch/single
Content-Type: application/json
```

**请求体**

```json
{
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "orders": [
    {
      "id": "order-001",
      "customer_id": "cust-001",
      "service_type": "cleaning",
      "service_date": "2026-01-11",
      "start_time": "09:00",
      "end_time": "11:00",
      "address": "北京市朝阳区xxx街道",
      "location": {
        "latitude": 39.9042,
        "longitude": 116.4074
      },
      "skills": ["cleaning"],
      "priority": 1
    }
  ],
  "employees": [
    {
      "id": "emp-001",
      "name": "李师傅",
      "skills": ["cleaning", "cooking"],
      "certifications": ["health_cert", "no_criminal_record"],
      "location": {
        "latitude": 39.9100,
        "longitude": 116.4200
      }
    }
  ],
  "customers": [
    {
      "id": "cust-001",
      "name": "王女士",
      "preferred_emp_ids": ["emp-001"],
      "blocked_emp_ids": []
    }
  ]
}
```

**响应**

```json
{
  "success": true,
  "results": [
    {
      "order_id": "order-001",
      "assigned_employee_id": "emp-001",
      "assigned_employee_name": "李师傅",
      "score": 92.5,
      "distance_km": 2.3,
      "travel_time_min": 15,
      "alternatives": [
        {
          "employee_id": "emp-002",
          "employee_name": "张师傅",
          "score": 85.0
        }
      ],
      "reason": "技能匹配、距离近、客户偏好"
    }
  ]
}
```

### 4.2 批量派单

**请求**

```http
POST /api/v1/dispatch/batch
Content-Type: application/json
```

请求体格式与单订单派单相同，但 `orders` 数组可包含多个订单。

### 4.3 路线优化

**请求**

```http
POST /api/v1/dispatch/route
Content-Type: application/json
```

**请求体**

```json
{
  "org_id": "550e8400-e29b-41d4-a716-446655440000",
  "orders": [...],
  "start_location": {
    "latitude": 39.9000,
    "longitude": 116.4000
  }
}
```

**响应**

```json
{
  "success": true,
  "route": [
    {"order_id": "order-003", "sequence": 1},
    {"order_id": "order-001", "sequence": 2},
    {"order_id": "order-002", "sequence": 3}
  ],
  "total_distance_km": 15.6,
  "estimated_time_min": 180
}
```

## 5. 护理计划API

### 5.1 创建护理计划

**请求**

```http
POST /api/v1/careplan/create
Content-Type: application/json
```

**请求体**

```json
{
  "customer_id": "cust-001",
  "care_plan": {
    "level": 3,
    "start_date": "2026-01-11",
    "end_date": "2027-01-10",
    "weekly_hours": 10,
    "service_items": [
      {
        "code": "basic_care",
        "name": "基础护理",
        "duration": 60,
        "frequency": "daily"
      },
      {
        "code": "health_check",
        "name": "健康检查",
        "duration": 30,
        "frequency": "weekly"
      }
    ],
    "frequency": "5_times_per_week"
  }
}
```

**响应**

```json
{
  "success": true,
  "data": {
    "id": "plan-001",
    "plan_no": "CP2026011101",
    "status": "active",
    ...
  }
}
```

### 5.2 生成服务订单

**请求**

```http
POST /api/v1/careplan/generate-orders
Content-Type: application/json
```

**请求体**

```json
{
  "care_plan": {...},
  "customer": {...},
  "period_start": "2026-01-11",
  "period_end": "2026-01-17"
}
```

### 5.3 推荐护理员

**请求**

```http
POST /api/v1/careplan/recommend-carers
Content-Type: application/json
```

## 6. 统计分析API

### 6.1 公平性分析

**请求**

```http
POST /api/v1/stats/fairness
Content-Type: application/json
```

**请求体**

```json
{
  "org_id": "...",
  "start_date": "2026-01-01",
  "end_date": "2026-01-31",
  "employees": [...],
  "assignments": [...]
}
```

**响应**

```json
{
  "success": true,
  "metrics": {
    "gini_coefficient": 0.15,
    "variance": 4.2,
    "workload_range": {"min": 35, "max": 45},
    "overall_score": 92.5,
    "employee_stats": [...]
  }
}
```

### 6.2 覆盖率分析

**请求**

```http
POST /api/v1/stats/coverage
```

### 6.3 工作量统计

**请求**

```http
POST /api/v1/stats/workload
```

## 7. 约束类型参考

### 7.1 通用约束

| 类型 | 说明 | 参数 |
|------|------|------|
| MaxHoursPerDay | 每日最大工时 | max_hours |
| MaxHoursPerWeek | 每周最大工时 | max_hours |
| MinRestBetweenShifts | 班次间最小休息时间 | min_hours |
| MaxConsecutiveDays | 最大连续工作天数 | max_days |
| SkillRequired | 技能要求 | skills |

### 7.2 餐饮约束

| 类型 | 说明 |
|------|------|
| PeakHoursMinStaff | 高峰时段最低人力 |
| PositionCoverage | 岗位覆盖 |
| SplitShift | 拆班支持 |

### 7.3 工厂约束

| 类型 | 说明 |
|------|------|
| ShiftRotationPattern | 班次轮换模式 |
| ProductionLineCoverage | 产线覆盖 |
| MaxConsecutiveNights | 最大连续夜班 |
| TeamTogether | 团队整体排班 |

### 7.4 派单约束

| 类型 | 说明 |
|------|------|
| ServiceAreaMatch | 服务区域匹配 |
| TravelTimeBuffer | 订单间缓冲时间 |
| MaxOrdersPerDay | 每日最大订单数 |
| CustomerPreference | 客户偏好 |
| CaregiverContinuity | 护理员连续性 |

## 8. 错误码

| 状态码 | 错误码 | 说明 |
|--------|--------|------|
| 400 | invalid_request | 请求参数错误 |
| 401 | unauthorized | 未授权 |
| 403 | forbidden | 权限不足 |
| 404 | not_found | 资源不存在 |
| 429 | rate_limit | 请求频率超限 |
| 500 | internal_error | 服务器内部错误 |

## 9. 最佳实践

### 9.1 性能优化

1. **批量请求**: 尽量使用批量接口减少请求次数
2. **合理分页**: 大量数据查询使用分页
3. **缓存结果**: 约束模板等静态数据可本地缓存
4. **异步处理**: 大规模排班考虑异步方式

### 9.2 错误处理

```python
import requests

def generate_schedule(data):
    try:
        resp = requests.post(
            "http://localhost:7012/api/v1/schedule/generate",
            json=data,
            headers={"Authorization": "Bearer pk_xxx"},
            timeout=60
        )
        resp.raise_for_status()
        return resp.json()
    except requests.exceptions.Timeout:
        # 处理超时
        pass
    except requests.exceptions.HTTPError as e:
        # 处理HTTP错误
        error = e.response.json()
        print(f"错误: {error['message']}")
```

### 9.3 安全建议

1. 不要在客户端暴露API密钥
2. 使用HTTPS加密传输
3. 定期轮换API密钥
4. 限制API密钥权限范围

## 10. SDK（示例）

### Python

```python
from paiban import PaibanClient

client = PaibanClient(
    api_key="pk_xxx",
    base_url="http://localhost:7012"
)

# 生成排班
result = client.schedule.generate(
    org_id="xxx",
    scenario="restaurant",
    employees=[...],
    shifts=[...]
)
```

### Go

```go
import "github.com/paiban/paiban-go"

client := paiban.NewClient("pk_xxx")

result, err := client.Schedule.Generate(context.Background(), &paiban.GenerateRequest{
    OrgID:    "xxx",
    Scenario: "restaurant",
    // ...
})
```

## 11. Webhook（规划中）

```json
{
  "event": "schedule.generated",
  "timestamp": "2026-01-11T10:30:00Z",
  "data": {
    "schedule_id": "sched-001",
    "org_id": "xxx",
    "status": "success"
  }
}
```

