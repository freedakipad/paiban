# PaiBan 排班引擎

> 通用智能排班引擎服务，支持餐饮、工厂、家政、长护险等多种场景

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## 🎯 功能特点

- **🔧 可配置约束系统** - 29种内置约束，硬约束/软约束分离，权重可调
- **🎯 智能排班生成** - 贪心算法 + 局部优化（禁忌搜索 + 模拟退火）
- **✅ 冲突检测验证** - 实时验证排班合规性，详细违规报告
- **📊 统计分析** - 工作量均衡、公平性评估、覆盖率分析
- **🔌 RESTful API** - 标准接口，易于集成
- **🌐 前端控制台** - 独立Web界面，可视化测试和配置
- **⏱️ 超时控制** - 支持请求超时，优雅降级返回部分结果
- **🔐 请求追踪** - Request ID 追踪，便于问题定位
- **🚦 速率限制** - Token Bucket 算法，保护服务稳定性

## 📦 支持场景

| 场景 | 特性 |
|------|------|
| 🍽️ **餐饮门店** | 高峰期排班、两头班、技能匹配、健康证要求 |
| 🏭 **工厂产线** | 三班倒、倒班模式、班组完整性、产线覆盖 |
| 🏠 **家政服务** | 派单优化、路线规划、客户偏好、服务区域 |
| 🏥 **长护险** | 护理计划、服务连续性、资质匹配、患者偏好 |

## 🚀 快速开始

### 环境要求

- Go 1.23+
- PostgreSQL 15+ (可选)
- Redis 6+ (可选)

### 快速启动

```bash
# 克隆项目
git clone https://github.com/freedakipad/paiban.git
cd paiban

# 一键启动
./scripts/quick-start.sh

# 或手动启动
go build -o bin/paiban cmd/server/main.go
./bin/paiban
```

### 验证服务

```bash
# 健康检查
curl http://localhost:7012/health

# API 信息
curl http://localhost:7012/api/v1/
```

服务默认端口：`7012`

📖 详细部署指南请参考：[部署文档](docs/deploy.md)

### 启动前端控制台

```bash
# 进入前端目录
cd frontend

# 启动静态服务器 (使用 Python)
python3 -m http.server 8888

# 访问 http://localhost:8888
```

## 🌐 前端控制台

PaiBan 提供独立的 Web 前端控制台，用于可视化测试和配置：

- **📋 场景预设** - 餐饮、工厂、家政、长护险一键切换
- **📝 业务视图** - 用业务语言解释请求和响应
- **📐 约束模板** - 查看和编辑各场景的约束配置
- **📚 约束库** - 浏览后端支持的全部29种约束
- **✏️ 在线编辑** - 修改约束参数，支持保存/取消
- **🗑️ 删除约束** - 从模板中移除不需要的约束
- **📥 从库添加** - 从约束库选择合适的约束添加到配置

## 📖 API 使用

### 服务端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/api/v1/` | GET | API 信息 |
| `/api/v1/schedule/generate` | POST | 生成排班 |
| `/api/v1/schedule/validate` | POST | 验证排班 |
| `/api/v1/constraints/templates` | GET | 获取约束模板 |
| `/api/v1/constraints/library` | GET | 获取约束库 |
| `/api/v1/stats/fairness` | POST | 公平性分析 |
| `/api/v1/stats/coverage` | POST | 覆盖率分析 |
| `/api/v1/stats/workload` | POST | 工作量统计 |
| `/api/v1/dispatch/single` | POST | 智能派单 |
| `/api/v1/dispatch/batch` | POST | 批量派单 |
| `/api/v1/dispatch/route` | POST | 最优路线 |
| `/metrics` | GET | Prometheus 指标 |

### 生成排班

```bash
curl -X POST http://localhost:7012/api/v1/schedule/generate \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: my-trace-id" \
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
      }
    ],
    "shifts": [
      {
        "id": "shift-001",
        "name": "早班",
        "code": "M",
        "start_time": "09:00",
        "end_time": "14:00",
        "duration": 300
      }
    ],
    "requirements": [
      {
        "shift_id": "shift-001",
        "date": "2024-01-15",
        "min_employees": 2,
        "position": "服务员",
        "note": "早班服务"
      }
    ],
    "options": {
      "timeout": 30,
      "optimization_level": "balanced",
      "consider_preferences": true
    }
  }'
```

### 获取约束库

```bash
curl http://localhost:7012/api/v1/constraints/library
```

返回示例：

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
        {"name": "max_hours", "type": "int", "default": "10", "min": "6", "max": "14"}
      ]
    }
  ]
}
```

## 📁 项目结构

```
paiban/
├── api/                    # API 定义 (OpenAPI)
├── cmd/
│   └── server/            # 主程序入口
├── configs/               # 配置文件
├── docs/                  # 文档
│   ├── design.md          # 设计文档
│   └── dev-test-plan.md   # 开发测试计划
├── frontend/              # 前端控制台
│   └── index.html         # 单页应用
├── internal/
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接
│   ├── handler/           # HTTP 处理器
│   ├── metrics/           # Prometheus 指标
│   └── repository/        # 数据访问层
├── pkg/
│   ├── errors/            # 统一错误处理
│   ├── logger/            # 日志框架 (zerolog)
│   ├── model/             # 数据模型
│   └── scheduler/         # 排班引擎核心
│       ├── constraint/    # 约束系统
│       │   └── builtin/   # 内置约束
│       ├── optimizer/     # 局部搜索优化
│       └── solver/        # 求解器
├── scripts/               # 脚本工具
└── tests/                 # 测试文件
```

## ⚙️ 约束系统

### 内置约束 (29种)

**硬约束（必须满足）：**

| 约束 | 代码 | 适用场景 |
|------|------|----------|
| 每日最大工时 | `max_hours_per_day` | 全部 |
| 每周最大工时 | `max_hours_per_week` | 全部 |
| 班次间最小休息 | `min_rest_between_shifts` | 全部 |
| 最大连续工作天数 | `max_consecutive_days` | 全部 |
| 技能与岗位匹配 | `skill_required` | 全部 |
| 行业资质认证 | `industry_certification` | 餐饮/家政/护理 |
| 倒班轮换规则 | `shift_rotation` | 工厂 |
| 最大连续夜班 | `max_consecutive_nights` | 工厂 |
| 产线24小时覆盖 | `production_line_coverage` | 工厂 |
| 服务区域匹配 | `service_area` | 家政/护理 |
| 服务时间窗口 | `time_window` | 家政/护理 |
| 护理资质等级 | `nursing_qualification` | 护理 |
| 每日最大服务患者数 | `max_patients_per_day` | 护理 |

**软约束（尽量满足）：**

| 约束 | 代码 | 适用场景 |
|------|------|----------|
| 工作量均衡 | `workload_balance` | 全部 |
| 员工偏好考虑 | `employee_preference` | 全部 |
| 减少加班 | `minimize_overtime` | 全部 |
| 高峰期人员覆盖 | `peak_hours_coverage` | 餐饮 |
| 两头班支持 | `split_shift` | 餐饮 |
| 岗位覆盖 | `position_coverage` | 餐饮 |
| 团队协作 | `team_together` | 工厂 |
| 路程时间优化 | `travel_time` | 家政/护理 |
| 服务连续性 | `service_continuity` | 护理 |

### 约束配置示例

```json
{
  "constraints": {
    "max_hours_per_day": 10,
    "max_hours_per_week": 44,
    "min_rest_between_shifts": 11,
    "max_consecutive_days": 6,
    "workload_balance_weight": 60,
    "preference_weight": 50,
    "minimize_overtime_weight": 70
  }
}
```

## 🔧 中间件功能

### 请求ID追踪

所有请求自动添加 `X-Request-ID` 响应头，支持链路追踪：

```bash
# 自定义 Request ID
curl -H "X-Request-ID: my-trace-123" http://localhost:7012/health
```

### 速率限制

使用 Token Bucket 算法，默认配置：
- 桶容量：100 请求
- 填充速率：10 请求/秒

### 超时控制

排班生成支持超时设置，超时后返回部分结果：

```json
{
  "options": {
    "timeout": 30
  }
}
```

### 优雅降级

当无法满足所有需求时，返回部分排班结果和未满足需求列表：

```json
{
  "assignments": [...],
  "partial": true,
  "unfilled": [
    {
      "shift_id": "...",
      "date": "2024-01-15",
      "position": "服务员",
      "required": 2,
      "assigned": 1,
      "shortage": 1,
      "reason": "员工不足"
    }
  ]
}
```

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行所有测试（详细输出）
go test -v ./...

# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 性能指标

| 场景 | 规模 | 响应时间 |
|------|------|----------|
| 100人/周 | 700 分配 | < 1s |
| 500人/周 | 3500 分配 | < 10s |
| 1000人/周 | 7000 分配 | < 30s |

### 优化措施

- FNV-1a 哈希算法替代字符串拼接
- 禁忌搜索 + 模拟退火混合算法
- 并发候选评估
- 结果缓存

## 🛠️ 开发

```bash
# 格式化代码
go fmt ./...

# 静态检查
golangci-lint run

# 编译
go build -o bin/paiban cmd/server/main.go

# 运行
./bin/paiban
```

## 📚 文档

- [API 使用说明](docs/api-usage.md) ⭐
- [详细 API 指南](docs/api-guide.md)
- [部署指南](docs/deploy.md)
- [设计文档](docs/design.md)
- [开发测试计划](docs/dev-test-plan.md)
- [API 规范](api/openapi.yaml)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
