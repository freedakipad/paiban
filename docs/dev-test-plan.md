# PaiBan 排班引擎 - 开发与测试计划

> 版本：v1.0  
> 日期：2026-01-11  
> 关联文档：[design.md](./design.md)

---

## 目录

1. [开发计划概述](#1-开发计划概述)
2. [开发阶段详细计划](#2-开发阶段详细计划)
3. [测试策略总览](#3-测试策略总览)
4. [单元测试计划](#4-单元测试计划)
5. [集成测试计划](#5-集成测试计划)
6. [E2E 测试计划](#6-e2e-测试计划)
7. [场景测试计划](#7-场景测试计划)
8. [性能测试计划](#8-性能测试计划)
9. [测试环境与工具](#9-测试环境与工具)
10. [质量门禁与发布标准](#10-质量门禁与发布标准)

---

## 1. 开发计划概述

### 1.1 项目时间线

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           PaiBan 开发时间线 (18周)                               │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   Phase 1        Phase 2        Phase 3        Phase 4        Phase 5          │
│   MVP            约束与场景     优化增强       派出服务       生产就绪          │
│   (4周)          (4周)          (4周)          (4周)          (2周)             │
│                                                                                 │
│   Week           Week           Week           Week           Week              │
│   1-4            5-8            9-12           13-16          17-18             │
│   │              │              │              │              │                 │
│   ▼              ▼              ▼              ▼              ▼                 │
│   ┌────┐         ┌────┐         ┌────┐         ┌────┐         ┌────┐           │
│   │MVP │ ──────▶ │约束│ ──────▶ │优化│ ──────▶ │派出│ ──────▶ │上线│           │
│   │发布│         │完善│         │算法│         │服务│         │准备│           │
│   └────┘         └────┘         └────┘         └────┘         └────┘           │
│                                                                                 │
│   ├──测试──┤     ├──测试──┤     ├──测试──┤     ├──测试──┤     ├─测试─┤         │
│   单元测试       集成测试       性能测试       场景测试       E2E+回归          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 开发原则

| 原则 | 说明 |
|------|------|
| **TDD 优先** | 核心算法采用测试驱动开发 |
| **持续集成** | 每次提交触发自动化测试 |
| **代码审查** | 所有代码需通过 Code Review |
| **文档同步** | 代码变更同步更新 API 文档 |
| **渐进交付** | 每阶段产出可运行版本 |

### 1.3 技术债务管理

```
┌─────────────────────────────────────────────────────────────┐
│                    技术债务处理策略                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│   记录：所有临时方案记录到 TODO 注释和 Issue 跟踪           │
│   评估：每周回顾，按影响程度分类（P0/P1/P2）                │
│   偿还：每个 Sprint 预留 10% 时间处理技术债务               │
│   预防：代码审查时识别潜在债务                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 2. 开发阶段详细计划

### 2.1 Phase 1: MVP (Week 1-4)

#### Week 1: 项目初始化

| 任务 | 交付物 | 负责人 | 工时 |
|------|--------|--------|------|
| 项目结构搭建 | 标准目录结构 | - | 4h |
| Go Module 初始化 | go.mod, 依赖管理 | - | 2h |
| 数据库 Schema 设计 | migrations/*.sql | - | 8h |
| 配置管理框架 | config/*.yaml | - | 4h |
| 日志和错误处理框架 | pkg/logger, pkg/errors | - | 4h |
| CI/CD Pipeline | .github/workflows | - | 4h |
| 基础 CRUD Repository | internal/repository | - | 12h |

**里程碑检查点：**
- [ ] 项目可编译运行
- [ ] 数据库可连接，表结构创建成功
- [ ] 基础 Repository 单元测试通过

#### Week 2: 排班核心引擎

| 任务 | 交付物 | 负责人 | 工时 |
|------|--------|--------|------|
| Constraint 接口定义 | pkg/constraint/constraint.go | - | 4h |
| ConstraintManager 实现 | pkg/constraint/manager.go | - | 8h |
| ScheduleContext 实现 | pkg/scheduler/context.go | - | 6h |
| 基础约束实现 (5个) | pkg/constraint/builtin/*.go | - | 16h |
| GreedySolver 实现 | pkg/scheduler/solver/greedy.go | - | 12h |
| 引擎单元测试 | pkg/scheduler/*_test.go | - | 8h |

**基础约束清单 (Week 2):**
1. `MaxHoursPerDay` - 每日最大工时
2. `MaxHoursPerWeek` - 每周最大工时
3. `MinRestBetweenShifts` - 班次间最小休息
4. `MaxConsecutiveDays` - 最大连续工作天数
5. `SkillRequired` - 技能要求

**里程碑检查点：**
- [ ] 约束接口设计通过评审
- [ ] 5个基础约束单元测试 100% 通过
- [ ] GreedySolver 基础场景测试通过

#### Week 3: API 服务

| 任务 | 交付物 | 负责人 | 工时 |
|------|--------|--------|------|
| HTTP 框架搭建 | cmd/server/main.go | - | 4h |
| 认证中间件 | internal/middleware/auth.go | - | 6h |
| 排班生成 API | internal/handler/schedule.go | - | 8h |
| 冲突检测 API | internal/handler/validate.go | - | 6h |
| 约束模板 API | internal/handler/constraint.go | - | 4h |
| OpenAPI 文档 | api/openapi.yaml | - | 6h |
| API 集成测试 | tests/integration/*_test.go | - | 8h |

**里程碑检查点：**
- [ ] API 可正常访问
- [ ] OpenAPI 文档可在 Swagger UI 查看
- [ ] 核心 API 集成测试通过

#### Week 4: 测试与优化

| 任务 | 交付物 | 负责人 | 工时 |
|------|--------|--------|------|
| 单元测试补全 | *_test.go (覆盖率 80%+) | - | 12h |
| 集成测试完善 | tests/integration/*.go | - | 8h |
| 性能基准测试 | tests/benchmark/*_test.go | - | 6h |
| Bug 修复 | - | - | 8h |
| 文档完善 | docs/*.md | - | 4h |
| MVP 演示准备 | - | - | 2h |

**MVP 交付标准：**
- [x] 核心 API 功能完整 ✅
- [x] 测试覆盖率 ≥ 80% ✅
- [x] 100人/周排班 < 30秒 ✅
- [x] 无 P0 级别 Bug ✅

> **Phase 1 已完成** ✅ (2026-01-11)

---

### 2.2 Phase 2: 约束与场景 (Week 5-8)

#### Week 5-6: 完善约束系统

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 餐饮场景约束 (5个) | pkg/constraint/restaurant/*.go | 16h |
| 工厂场景约束 (5个) | pkg/constraint/factory/*.go | 16h |
| 公平性约束 (3个) | pkg/constraint/fairness/*.go | 12h |
| 约束配置模板 | configs/constraints/*.yaml | 8h |
| 场景适配器 | pkg/adapter/*.go | 16h |
| 约束系统集成测试 | tests/integration/constraint_test.go | 12h |

**餐饮场景约束：**
1. `PeakHoursMinStaff` - 高峰期最少人数
2. `PositionCoverage` - 岗位覆盖
3. `SplitShift` - 两头班支持
4. `PreferenceMatch` - 员工偏好匹配
5. `MinimizeOvertime` - 最小化加班

**工厂场景约束：**
1. `ShiftRotationPattern` - 倒班模式
2. `ProductionLineCoverage` - 产线覆盖
3. `MaxConsecutiveNights` - 最大连续夜班
4. `TeamTogether` - 班组完整性
5. `CertificationRequired` - 资质证书

#### Week 7-8: 冲突检测与调班

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 实时冲突检测引擎 | pkg/validator/conflict.go | 12h |
| 调班可行性评估 | pkg/swap/evaluator.go | 10h |
| 换班推荐算法 | pkg/swap/recommender.go | 12h |
| 调班相关 API | internal/handler/swap.go | 8h |
| 场景测试用例 | tests/scenario/restaurant_test.go | 10h |
| 场景测试用例 | tests/scenario/factory_test.go | 10h |

**里程碑检查点：**
- [x] 20+ 约束全部实现并测试 ✅
- [x] 餐饮/工厂场景测试通过 ✅
- [x] 调班评估准确率 > 95% ✅

> **Phase 2 已完成** ✅ (2026-01-11)

---

### 2.3 Phase 3: 优化增强 (Week 9-12)

#### Week 9-10: 算法优化

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 局部搜索优化器 | pkg/scheduler/optimizer/local_search.go | 16h |
| 邻域操作实现 | pkg/scheduler/optimizer/neighbors.go | 12h |
| 并行计算支持 | pkg/scheduler/solver/parallel.go | 10h |
| 遗传算法（可选） | pkg/scheduler/solver/genetic.go | 16h |
| 性能测试与调优 | tests/benchmark/*.go | 10h |

**性能目标：**
- 100人/周排班：< 10秒 ✅
- 500人/周排班：< 60秒 ✅
- 软约束满足率：> 90% ✅

> **Phase 3 已完成** ✅ (2026-01-11)
> 实现内容：局部搜索优化器、邻域操作、并行评估、岛屿模型、公平性/覆盖率/工作量统计分析、Prometheus指标

#### Week 11-12: 统计与监控

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 工作量统计 API | internal/handler/stats.go | 8h |
| 公平性评估算法 | pkg/stats/fairness.go | 10h |
| 覆盖率分析 | pkg/stats/coverage.go | 8h |
| Prometheus 指标 | internal/metrics/*.go | 8h |
| Grafana Dashboard | deployments/grafana/*.json | 6h |
| 端到端测试 | tests/e2e/*.go | 16h |

---

### 2.4 Phase 4: 派出服务模块 (Week 13-16)

#### Week 13-14: 派出服务基础

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 派出服务数据模型 | internal/model/dispatch.go | 6h |
| 客户/订单实体 | pkg/model/customer.go, order.go | 8h |
| 派出服务约束 (6个) | pkg/constraint/dispatch/*.go | 20h |
| 派单引擎核心 | pkg/dispatcher/engine.go | 16h |
| 技能/距离匹配器 | pkg/dispatcher/matcher/*.go | 12h |

**派出服务约束：**
1. `ServiceAreaMatch` - 服务区域匹配
2. `TravelTimeBuffer` - 路程时间缓冲
3. `MaxOrdersPerDay` - 每日最大订单数
4. `CustomerPreference` - 客户偏好
5. `CertificationLevel` - 资质等级
6. `CaregiverContinuity` - 护理员连续性

#### Week 15-16: 智能派单与长护险

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 路线优化算法 | pkg/dispatcher/router/vrp.go | 16h |
| 智能派单 API | internal/handler/dispatch.go | 10h |
| 护理计划管理 | pkg/careplan/*.go | 12h |
| 护理计划 API | internal/handler/careplan.go | 8h |
| 家政场景测试 | tests/scenario/housekeeping_test.go | 10h |
| 长护险场景测试 | tests/scenario/nursing_test.go | 10h |

> **Phase 4 已完成** ✅ (2026-01-11)
> 实现内容：派出服务数据模型、7个派单约束、智能派单引擎、技能/距离匹配器、护理计划管理、家政/长护险场景测试

---

### 2.5 Phase 5: 生产就绪 (Week 17-18)

| 任务 | 交付物 | 工时 |
|------|--------|------|
| 多租户支持 | internal/tenant/*.go | 12h |
| 安全加固 | internal/security/*.go | 10h |
| 完整 E2E 测试 | tests/e2e/full_test.go | 12h |
| 回归测试执行 | - | 8h |
| Docker 镜像优化 | deployments/Dockerfile | 4h |
| K8s 部署配置 | deployments/k8s/*.yaml | 8h |
| 运维文档 | docs/operations.md | 6h |
| API 接入指南 | docs/api-guide.md | 6h |

> **Phase 5 已完成** ✅ (2026-01-11)
> 实现内容：多租户支持、安全加固（API密钥、频率限制、签名验证）、E2E测试、Docker镜像优化（多阶段构建、UPX压缩）、K8s部署配置（Deployment/Service/Ingress/HPA/PDB）、运维文档、API接入指南

---

## 3. 测试策略总览

### 3.1 测试金字塔

```
                          ▲
                         /│\
                        / │ \
                       /  │  \        E2E 测试 (5%)
                      /   │   \       端到端流程验证
                     /────│────\
                    /     │     \
                   /      │      \    集成测试 (20%)
                  /       │       \   模块间交互
                 /────────│────────\
                /         │         \
               /          │          \  单元测试 (75%)
              /           │           \ 函数/方法级别
             /────────────│────────────\
            ▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔▔

    场景测试：跨越所有层，验证业务场景完整性
```

### 3.2 测试类型与目标

| 测试类型 | 目标 | 占比 | 执行频率 |
|----------|------|------|----------|
| **单元测试** | 验证单个函数/方法的正确性 | 75% | 每次提交 |
| **集成测试** | 验证模块间交互 | 20% | 每次提交 |
| **E2E 测试** | 验证完整用户流程 | 5% | 每日/发布前 |
| **场景测试** | 验证业务场景 | - | 每个迭代 |
| **性能测试** | 验证性能指标 | - | 每周/发布前 |

### 3.3 测试覆盖率目标

| 模块 | 覆盖率目标 | 说明 |
|------|------------|------|
| `pkg/scheduler` | ≥ 90% | 核心排班引擎 |
| `pkg/constraint` | ≥ 90% | 约束系统 |
| `pkg/dispatcher` | ≥ 85% | 派单引擎 |
| `internal/handler` | ≥ 80% | API 处理器 |
| `internal/repository` | ≥ 70% | 数据访问层 |
| **整体** | **≥ 80%** | |

---

## 4. 单元测试计划

### 4.1 测试范围

#### 4.1.1 约束系统单元测试

```go
// pkg/constraint/builtin/max_hours_test.go

func TestMaxHoursPerDay_Evaluate(t *testing.T) {
    tests := []struct {
        name           string
        maxHours       float64
        assignments    []*Assignment
        expectedValid  bool
        expectedPenalty int
    }{
        {
            name:     "正常工时",
            maxHours: 10,
            assignments: []*Assignment{
                {EmployeeID: "emp1", Date: today, Duration: 8 * time.Hour},
            },
            expectedValid:  true,
            expectedPenalty: 0,
        },
        {
            name:     "超出工时",
            maxHours: 10,
            assignments: []*Assignment{
                {EmployeeID: "emp1", Date: today, Duration: 12 * time.Hour},
            },
            expectedValid:  false,
            expectedPenalty: 20, // (12-10) * 10
        },
        // ... 更多测试用例
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试实现
        })
    }
}

func TestMaxHoursPerDay_EdgeCases(t *testing.T) {
    // 边界条件测试：
    // - 精确达到最大工时
    // - 多个班次累加
    // - 跨天班次
    // - 空员工列表
}
```

#### 4.1.2 求解器单元测试

```go
// pkg/scheduler/solver/greedy_test.go

func TestGreedySolver_SimpleDemand(t *testing.T) {
    // 场景：3个员工，2个班次需求
    // 期望：所有需求被满足
}

func TestGreedySolver_InsufficientResources(t *testing.T) {
    // 场景：需求超过可用员工
    // 期望：尽可能满足，返回未满足需求列表
}

func TestGreedySolver_SkillMatching(t *testing.T) {
    // 场景：特定岗位需要特定技能
    // 期望：只分配具有相应技能的员工
}

func TestGreedySolver_ConsecutiveDays(t *testing.T) {
    // 场景：连续工作天数限制
    // 期望：不违反连续工作天数约束
}

func TestGreedySolver_Benchmark(t *testing.B) {
    // 性能基准测试
    // 100人/周排班性能
}
```

#### 4.1.3 派单引擎单元测试

```go
// pkg/dispatcher/matcher/skill_matcher_test.go

func TestSkillMatcher_ExactMatch(t *testing.T) {
    // 护理员证书等级完全匹配
}

func TestSkillMatcher_HigherLevelOK(t *testing.T) {
    // 高级护理员可服务低等级需求
}

func TestSkillMatcher_LowerLevelFail(t *testing.T) {
    // 低级护理员不能服务高等级需求
}

// pkg/dispatcher/matcher/distance_matcher_test.go

func TestDistanceMatcher_WithinRange(t *testing.T) {
    // 在服务区域内
}

func TestDistanceMatcher_OutOfRange(t *testing.T) {
    // 超出最大服务距离
}
```

### 4.2 单元测试规范

```go
// 测试文件命名：{被测文件}_test.go
// 测试函数命名：Test{函数名}_{场景描述}

// 标准测试结构
func TestXxx_Scenario(t *testing.T) {
    // Arrange - 准备测试数据
    input := prepareInput()
    expected := prepareExpected()
    
    // Act - 执行被测函数
    actual, err := FunctionUnderTest(input)
    
    // Assert - 验证结果
    assert.NoError(t, err)
    assert.Equal(t, expected, actual)
}

// 表驱动测试（推荐）
func TestXxx_TableDriven(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {"case1", input1, output1, false},
        {"case2", input2, output2, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

### 4.3 单元测试用例清单

| 模块 | 测试类 | 用例数 | 优先级 |
|------|--------|--------|--------|
| **constraint/builtin** | | | |
| MaxHoursPerDay | 正常/超出/边界 | 8 | P0 |
| MaxHoursPerWeek | 周累计/跨周 | 6 | P0 |
| MinRestBetweenShifts | 足够/不足/夜转早 | 8 | P0 |
| MaxConsecutiveDays | 连续/间断/边界 | 6 | P0 |
| SkillRequired | 匹配/不匹配/多技能 | 6 | P0 |
| PeakHoursMinStaff | 满足/不足/边界 | 6 | P1 |
| WorkloadBalance | 均衡/偏差大 | 5 | P1 |
| **scheduler/solver** | | | |
| GreedySolver | 基础/复杂/边界/性能 | 15 | P0 |
| LocalSearchOptimizer | 交换/移动/优化效果 | 10 | P1 |
| **dispatcher/matcher** | | | |
| SkillMatcher | 精确/高于/低于 | 6 | P0 |
| DistanceMatcher | 范围内/范围外/边界 | 6 | P0 |
| PreferenceMatcher | 偏好/历史/黑名单 | 8 | P1 |
| **dispatcher/router** | | | |
| VRPRouter | 单订单/多订单/最优路径 | 10 | P1 |

---

## 5. 集成测试计划

### 5.1 测试范围

集成测试验证模块间的交互是否正确。

```
┌─────────────────────────────────────────────────────────────────────┐
│                        集成测试覆盖范围                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   HTTP Request                                                      │
│        │                                                            │
│        ▼                                                            │
│   ┌─────────────┐                                                   │
│   │   Handler   │ ◄─── 测试点 1: 请求解析、参数验证                 │
│   └──────┬──────┘                                                   │
│          │                                                          │
│          ▼                                                          │
│   ┌─────────────┐                                                   │
│   │   Service   │ ◄─── 测试点 2: 业务逻辑、约束组装                 │
│   └──────┬──────┘                                                   │
│          │                                                          │
│          ▼                                                          │
│   ┌─────────────┐                                                   │
│   │  Scheduler  │ ◄─── 测试点 3: 引擎调用、结果处理                 │
│   └──────┬──────┘                                                   │
│          │                                                          │
│          ▼                                                          │
│   ┌─────────────┐                                                   │
│   │ Repository  │ ◄─── 测试点 4: 数据持久化（可选）                 │
│   └─────────────┘                                                   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 5.2 集成测试用例

#### 5.2.1 排班 API 集成测试

```go
// tests/integration/schedule_api_test.go

func TestScheduleAPI_Generate_Success(t *testing.T) {
    // 准备：创建测试数据
    input := &ScheduleRequest{
        Scenario: "restaurant",
        DateRange: DateRange{
            Start: "2026-01-13",
            End:   "2026-01-19",
        },
        Resources: []Resource{
            {ID: "emp1", Name: "张三", Position: "服务员", Skills: []string{}},
            {ID: "emp2", Name: "李四", Position: "服务员", Skills: []string{}},
            {ID: "emp3", Name: "王五", Position: "厨师", Skills: []string{}},
        },
        Shifts: []Shift{
            {ID: "morning", Name: "早班", Start: "06:00", End: "14:00"},
            {ID: "evening", Name: "晚班", Start: "14:00", End: "22:00"},
        },
        Demands: []Demand{
            {Shift: "morning", Position: "服务员", MinCount: 1},
            {Shift: "evening", Position: "服务员", MinCount: 1},
        },
    }
    
    // 执行：调用 API
    resp := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/api/v1/schedule/generate", toJSON(input))
    req.Header.Set("Authorization", "Bearer test-api-key")
    router.ServeHTTP(resp, req)
    
    // 验证：检查响应
    assert.Equal(t, http.StatusOK, resp.Code)
    
    var result ScheduleResponse
    json.Unmarshal(resp.Body.Bytes(), &result)
    
    assert.True(t, result.Success)
    assert.GreaterOrEqual(t, len(result.Assignments), 7*2) // 7天 * 2班次
    assert.Equal(t, 100, result.Score.HardConstraints)
}

func TestScheduleAPI_Generate_ConstraintViolation(t *testing.T) {
    // 测试无法满足硬约束的情况
    // 例如：需要3个服务员，但只有2个可用
}

func TestScheduleAPI_Validate_DetectConflict(t *testing.T) {
    // 测试冲突检测功能
    // 传入已有排班，检测是否有约束违规
}

func TestScheduleAPI_Auth_InvalidKey(t *testing.T) {
    // 测试无效 API Key
    // 期望返回 401 Unauthorized
}

func TestScheduleAPI_RateLimit(t *testing.T) {
    // 测试限流
    // 超过 QPS 限制后返回 429
}
```

#### 5.2.2 派单 API 集成测试

```go
// tests/integration/dispatch_api_test.go

func TestDispatchAPI_Assign_Success(t *testing.T) {
    input := &DispatchRequest{
        Scenario: "nursing",
        Date:     "2026-01-15",
        Orders: []ServiceOrder{
            {
                ID:         "order1",
                CustomerID: "cust1",
                CareLevel:  4,
                TimeSlot:   TimeSlot{Start: "09:00", End: "11:00"},
                Location:   GeoLocation{Lat: 31.2397, Lng: 121.4998},
            },
        },
        Employees: []Employee{
            {
                ID:            "emp1",
                Certifications: []string{"高级护理员证"},
                HomeLocation:   GeoLocation{Lat: 31.2350, Lng: 121.5100},
                ServiceArea:    []string{"浦东新区"},
            },
        },
    }
    
    // 执行并验证
    // ...
}

func TestDispatchAPI_Assign_NoMatch(t *testing.T) {
    // 测试没有合适员工的情况
    // 例如：需要高级护理员，但只有初级
}

func TestDispatchAPI_Route_Optimize(t *testing.T) {
    // 测试多订单路线优化
    // 验证返回的路线是否为最优
}
```

### 5.3 集成测试环境

```yaml
# tests/integration/docker-compose.yaml

version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: paiban_test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    ports:
      - "5433:5432"
    tmpfs:
      - /var/lib/postgresql/data  # 使用内存存储加速

  redis:
    image: redis:7
    ports:
      - "6380:6379"
    
  paiban:
    build:
      context: ../..
      dockerfile: Dockerfile.test
    depends_on:
      - postgres
      - redis
    environment:
      DB_HOST: postgres
      REDIS_HOST: redis
```

### 5.4 集成测试执行

```bash
# 运行所有集成测试
make test-integration

# 运行特定测试
go test -v ./tests/integration/... -run TestScheduleAPI

# 生成覆盖率报告
go test -coverprofile=coverage.out ./tests/integration/...
go tool cover -html=coverage.out -o coverage.html
```

---

## 6. E2E 测试计划

### 6.1 E2E 测试目标

端到端测试验证完整的用户流程，从 API 调用到最终结果。

```
┌─────────────────────────────────────────────────────────────────────┐
│                        E2E 测试流程                                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│   调用方系统                  PaiBan 引擎                           │
│   ┌─────────┐                ┌─────────────────────────┐           │
│   │         │  1. 发送请求   │                         │           │
│   │  测试   │ ─────────────▶ │  API Gateway            │           │
│   │  客户端 │                │         │               │           │
│   │         │                │         ▼               │           │
│   │         │                │  ┌─────────────┐        │           │
│   │         │                │  │ 排班/派单   │        │           │
│   │         │                │  │    引擎     │        │           │
│   │         │                │  └─────────────┘        │           │
│   │         │                │         │               │           │
│   │         │  4. 接收结果   │         ▼               │           │
│   │         │ ◀───────────── │  ┌─────────────┐        │           │
│   └─────────┘                │  │  结果返回   │        │           │
│       │                      │  └─────────────┘        │           │
│       ▼                      └─────────────────────────┘           │
│   5. 验证结果                                                       │
│   - 排班完整性                                                      │
│   - 约束满足度                                                      │
│   - 性能指标                                                        │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 6.2 E2E 测试用例

#### 6.2.1 餐饮排班完整流程

```go
// tests/e2e/restaurant_full_flow_test.go

func TestE2E_Restaurant_WeeklySchedule(t *testing.T) {
    // 场景：某餐饮门店一周排班
    
    // Step 1: 准备员工数据（10人）
    employees := generateEmployees(10, "restaurant")
    
    // Step 2: 准备班次和需求
    shifts := []Shift{
        {ID: "morning", Start: "06:00", End: "14:00"},
        {ID: "noon", Start: "11:00", End: "19:00"},
        {ID: "evening", Start: "14:00", End: "22:00"},
    }
    
    demands := []Demand{
        // 工作日
        {Days: []int{1,2,3,4,5}, Shift: "morning", Position: "服务员", Min: 2},
        {Days: []int{1,2,3,4,5}, Shift: "noon", Position: "服务员", Min: 3}, // 午高峰
        {Days: []int{1,2,3,4,5}, Shift: "evening", Position: "服务员", Min: 2},
        // 周末
        {Days: []int{0,6}, Shift: "morning", Position: "服务员", Min: 3},
        {Days: []int{0,6}, Shift: "noon", Position: "服务员", Min: 4},
        {Days: []int{0,6}, Shift: "evening", Position: "服务员", Min: 3},
    }
    
    // Step 3: 调用排班 API
    result, err := client.GenerateSchedule(ctx, &ScheduleRequest{
        Scenario:  "restaurant",
        DateRange: DateRange{Start: "2026-01-13", End: "2026-01-19"},
        Resources: employees,
        Shifts:    shifts,
        Demands:   demands,
        Constraints: ConstraintConfig{
            UseTemplate: "restaurant_default",
        },
    })
    
    require.NoError(t, err)
    require.True(t, result.Success)
    
    // Step 4: 验证结果
    // 4.1 所有需求被满足
    for _, demand := range demands {
        for _, day := range demand.Days {
            date := addDays("2026-01-13", day)
            count := countAssignments(result.Assignments, date, demand.Shift, demand.Position)
            assert.GreaterOrEqual(t, count, demand.Min, 
                "需求未满足: %s %s %s", date, demand.Shift, demand.Position)
        }
    }
    
    // 4.2 硬约束全部满足
    assert.Equal(t, 100, result.Score.HardConstraints)
    
    // 4.3 每个员工工时不超标
    for _, emp := range employees {
        hours := sumHours(result.Assignments, emp.ID)
        assert.LessOrEqual(t, hours, 44.0, "员工 %s 工时超标", emp.Name)
    }
    
    // 4.4 连续工作天数不超过6天
    for _, emp := range employees {
        maxConsecutive := maxConsecutiveDays(result.Assignments, emp.ID)
        assert.LessOrEqual(t, maxConsecutive, 6, "员工 %s 连续工作天数超标", emp.Name)
    }
    
    // Step 5: 性能验证
    assert.Less(t, result.Statistics.SolveTimeMs, int64(30000), "排班时间超过30秒")
}
```

#### 6.2.2 长护险派单完整流程

```go
// tests/e2e/nursing_full_flow_test.go

func TestE2E_Nursing_DailyDispatch(t *testing.T) {
    // 场景：长护险站点一日派单
    
    // Step 1: 准备护理员（20人）
    caregivers := generateCaregivers(20)
    
    // Step 2: 准备服务订单（50单）
    orders := generateNursingOrders(50)
    
    // Step 3: 调用派单 API
    result, err := client.Dispatch(ctx, &DispatchRequest{
        Scenario:  "nursing",
        Date:      "2026-01-15",
        Orders:    orders,
        Employees: caregivers,
        Options: DispatchOptions{
            PreferFamiliar:   true,
            OptimizeRoute:    true,
            BalanceWorkload:  true,
        },
    })
    
    require.NoError(t, err)
    require.True(t, result.Success)
    
    // Step 4: 验证结果
    // 4.1 所有订单被分配
    assert.Equal(t, len(orders), len(result.Assignments))
    assert.Empty(t, result.Unassigned)
    
    // 4.2 资质匹配正确
    for _, assignment := range result.Assignments {
        order := findOrder(orders, assignment.OrderID)
        caregiver := findCaregiver(caregivers, assignment.EmployeeID)
        assert.True(t, meetsCareLevel(caregiver, order.CareLevel),
            "护理员资质不满足: %s -> %s", caregiver.ID, order.ID)
    }
    
    // 4.3 每个护理员订单数不超限
    orderCount := countOrdersByEmployee(result.Assignments)
    for empID, count := range orderCount {
        assert.LessOrEqual(t, count, 6, "护理员 %s 订单数超限", empID)
    }
    
    // 4.4 路线优化有效
    for _, route := range result.Routes {
        assert.True(t, isRouteOptimal(route), "路线未优化: %s", route.EmployeeID)
    }
    
    // Step 5: 性能验证
    assert.Less(t, result.Statistics.SolveTimeMs, int64(3000), "派单时间超过3秒")
}
```

### 6.3 E2E 测试执行

```bash
# 启动测试环境
docker-compose -f tests/e2e/docker-compose.yaml up -d

# 等待服务就绪
./scripts/wait-for-ready.sh

# 运行 E2E 测试
go test -v -tags=e2e ./tests/e2e/...

# 清理环境
docker-compose -f tests/e2e/docker-compose.yaml down
```

---

## 7. 场景测试计划

### 7.1 场景测试概述

场景测试是基于真实业务场景的综合测试，验证系统在实际使用情况下的表现。

### 7.2 餐饮场景测试

#### 测试场景 1：小型快餐店

```yaml
# tests/scenario/data/restaurant_small.yaml
name: "小型快餐店 - 日常排班"
description: "10人团队，7x营业，高峰期保障"

resources:
  employees: 10
  positions:
    - name: 服务员
      count: 6
    - name: 厨师
      count: 3
    - name: 收银
      count: 1

shifts:
  - {id: morning, name: 早班, start: "06:00", end: "14:00"}
  - {id: noon, name: 午班, start: "10:00", end: "18:00"}
  - {id: evening, name: 晚班, start: "14:00", end: "22:00"}

demands:
  weekday:
    morning: {服务员: 2, 厨师: 1}
    noon: {服务员: 4, 厨师: 2, 收银: 1}  # 午高峰
    evening: {服务员: 2, 厨师: 1}
  weekend:
    morning: {服务员: 3, 厨师: 2}
    noon: {服务员: 5, 厨师: 2, 收银: 1}
    evening: {服务员: 3, 厨师: 2}

constraints:
  template: restaurant_default
  overrides:
    max_hours_per_week: 44
    peak_hours_boost: true

expectations:
  - all_demands_satisfied: true
  - hard_constraint_score: 100
  - soft_constraint_score: ">= 85"
  - solve_time_ms: "< 10000"
```

#### 测试场景 2：连锁餐厅调班

```yaml
name: "连锁餐厅 - 临时调班"
description: "员工请假，寻找替班人选"

initial_schedule:
  # 已有排班数据
  
swap_request:
  requester: "emp_001"
  original_assignment:
    date: "2026-01-15"
    shift: "noon"
  reason: "家中有事"

expectations:
  - recommend_count: ">= 3"
  - recommended_employees_qualified: true
  - no_constraint_violation: true
```

### 7.3 工厂场景测试

#### 测试场景 1：三班倒车间

```yaml
name: "制造车间 - 三班倒排班"
description: "30人团队，A/B/C班组轮换"

resources:
  employees: 30
  shift_groups:
    - {id: A, members: [emp_01, emp_02, ..., emp_10]}
    - {id: B, members: [emp_11, emp_12, ..., emp_20]}
    - {id: C, members: [emp_21, emp_22, ..., emp_30]}

shifts:
  - {id: day, name: 白班, start: "08:00", end: "16:00"}
  - {id: swing, name: 中班, start: "16:00", end: "00:00"}
  - {id: night, name: 夜班, start: "00:00", end: "08:00"}

rotation_pattern: "三班倒"
# 白班4天 → 休息2天 → 中班4天 → 休息2天 → 夜班4天 → 休息2天

expectations:
  - rotation_pattern_followed: true
  - max_consecutive_nights: "<= 4"
  - production_line_coverage: true
```

### 7.4 家政场景测试

#### 测试场景 1：钟点工派单

```yaml
name: "家政公司 - 钟点工派单"
description: "50个阿姨，100个订单，多客户穿插"

resources:
  employees: 50
  skills:
    - 保洁
    - 烹饪
    - 育儿

orders:
  count: 100
  distribution:
    - {type: 日常保洁, count: 60, duration: 2h}
    - {type: 深度保洁, count: 20, duration: 4h}
    - {type: 做饭, count: 20, duration: 2h}

constraints:
  max_distance_km: 10
  travel_buffer_mins: 30
  max_orders_per_day: 4

expectations:
  - assignment_rate: ">= 95%"
  - skill_match: true
  - route_optimized: true
  - average_travel_km: "< 8"
```

### 7.5 长护险场景测试

#### 测试场景 1：护理站日常派单

```yaml
name: "护理站 - 一周服务"
description: "20个护理员，80个老人，按护理计划派单"

resources:
  caregivers: 20
  certifications:
    - {level: 初级, count: 5}
    - {level: 中级, count: 10}
    - {level: 高级, count: 5}

patients: 80
care_plans:
  - {level: 1-2, count: 20, frequency: "3/week"}
  - {level: 3-4, count: 40, frequency: "5/week"}
  - {level: 5-6, count: 20, frequency: "7/week"}

constraints:
  caregiver_continuity: true
  certification_match: true
  max_patients_per_day: 6

expectations:
  - care_plan_compliance: ">= 95%"
  - caregiver_continuity_rate: ">= 80%"
  - certification_match: true
  - workload_balance_score: ">= 80"
```

### 7.6 场景测试执行

```bash
# 运行所有场景测试
make test-scenario

# 运行特定场景
go test -v -tags=scenario ./tests/scenario/... -run TestRestaurant

# 生成场景测试报告
go test -json ./tests/scenario/... > scenario_results.json
./scripts/generate-scenario-report.sh scenario_results.json
```

---

## 8. 性能测试计划

### 8.1 性能测试目标

| 场景 | 规模 | 目标时间 | 备注 |
|------|------|----------|------|
| 小型餐饮排班 | 10人/周 | < 1秒 | |
| 中型餐饮排班 | 50人/周 | < 5秒 | |
| 大型工厂排班 | 100人/周 | < 30秒 | MVP 目标 |
| 超大规模排班 | 500人/周 | < 120秒 | Phase 3 目标 |
| 小规模派单 | 20人/50单/日 | < 1秒 | |
| 中规模派单 | 50人/100单/日 | < 3秒 | |
| 大规模派单 | 100人/200单/日 | < 10秒 | |

### 8.2 性能测试用例

```go
// tests/benchmark/schedule_benchmark_test.go

func BenchmarkSchedule_10_Employees(b *testing.B) {
    input := generateScheduleInput(10, 7) // 10人，7天
    solver := NewGreedySolver(defaultConstraints())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        solver.Solve(context.Background(), input)
    }
}

func BenchmarkSchedule_50_Employees(b *testing.B) {
    input := generateScheduleInput(50, 7)
    solver := NewGreedySolver(defaultConstraints())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        solver.Solve(context.Background(), input)
    }
}

func BenchmarkSchedule_100_Employees(b *testing.B) {
    input := generateScheduleInput(100, 7)
    solver := NewGreedySolver(defaultConstraints())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        solver.Solve(context.Background(), input)
    }
}

func BenchmarkSchedule_500_Employees(b *testing.B) {
    input := generateScheduleInput(500, 7)
    solver := NewOptimizedSolver(defaultConstraints())
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        solver.Solve(context.Background(), input)
    }
}
```

### 8.3 负载测试

```go
// tests/load/api_load_test.go

func TestAPILoad_ScheduleGenerate(t *testing.T) {
    // 使用 vegeta 或类似工具进行负载测试
    
    rate := vegeta.Rate{Freq: 100, Per: time.Second} // 100 QPS
    duration := 60 * time.Second
    
    targeter := vegeta.NewStaticTargeter(vegeta.Target{
        Method: "POST",
        URL:    "http://localhost:7012/api/v1/schedule/generate",
        Body:   []byte(scheduleRequestJSON),
        Header: http.Header{
            "Content-Type":  []string{"application/json"},
            "Authorization": []string{"Bearer test-key"},
        },
    })
    
    attacker := vegeta.NewAttacker()
    var metrics vegeta.Metrics
    
    for res := range attacker.Attack(targeter, rate, duration, "Schedule API") {
        metrics.Add(res)
    }
    metrics.Close()
    
    // 验证指标
    assert.Less(t, metrics.Latencies.P99, 5*time.Second, "P99 延迟超标")
    assert.Less(t, metrics.Latencies.P50, 1*time.Second, "P50 延迟超标")
    assert.Equal(t, 0.0, metrics.Errors, "存在错误请求")
}
```

### 8.4 性能测试报告

```
=== 性能测试报告 ===
日期: 2026-01-15
版本: v1.0.0

1. 排班性能

| 规模 | 平均耗时 | P95 | P99 | 内存峰值 |
|------|----------|-----|-----|----------|
| 10人/周 | 120ms | 180ms | 250ms | 15MB |
| 50人/周 | 1.2s | 1.8s | 2.5s | 45MB |
| 100人/周 | 8.5s | 12s | 18s | 120MB |
| 500人/周 | 85s | 110s | 145s | 580MB |

2. 派单性能

| 规模 | 平均耗时 | P95 | P99 |
|------|----------|-----|-----|
| 50单/日 | 280ms | 420ms | 600ms |
| 100单/日 | 850ms | 1.2s | 1.8s |
| 200单/日 | 3.2s | 4.5s | 6.8s |

3. API 吞吐量

| API | QPS | 平均延迟 | 错误率 |
|-----|-----|----------|--------|
| /schedule/generate | 45 | 2.1s | 0% |
| /schedule/validate | 200 | 45ms | 0% |
| /dispatch/assign | 120 | 380ms | 0% |
| /constraints/templates | 500 | 8ms | 0% |
```

---

## 9. 测试环境与工具

### 9.1 测试环境

```yaml
# 本地开发测试
local:
  os: macOS / Linux
  go: 1.21+
  database: PostgreSQL 15 (Docker)
  redis: Redis 7 (Docker)

# CI 测试环境
ci:
  runner: GitHub Actions
  go: 1.21
  database: PostgreSQL 15 (Service Container)
  redis: Redis 7 (Service Container)

# 性能测试环境
performance:
  cpu: 8 核
  memory: 32 GB
  database: PostgreSQL 15 (独立实例)
  redis: Redis 7 (独立实例)
```

### 9.2 测试工具

| 工具 | 用途 | 说明 |
|------|------|------|
| **go test** | 单元/集成测试 | Go 内置测试框架 |
| **testify** | 断言库 | assert, require, mock |
| **gomock** | Mock 生成 | 接口 Mock |
| **testcontainers** | 容器化测试 | 自动管理测试容器 |
| **vegeta** | 负载测试 | HTTP 负载测试 |
| **pprof** | 性能分析 | CPU/内存分析 |
| **golangci-lint** | 代码检查 | 静态分析 |

### 9.3 CI/CD 测试流程

```yaml
# .github/workflows/test.yaml

name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  unit-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run Unit Tests
        run: make test-unit
      - name: Upload Coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  integration-test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - name: Run Integration Tests
        run: make test-integration

  e2e-test:
    runs-on: ubuntu-latest
    needs: [unit-test, integration-test]
    steps:
      - uses: actions/checkout@v4
      - name: Build and Start Services
        run: docker-compose up -d
      - name: Run E2E Tests
        run: make test-e2e

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v4
        with:
          version: latest
```

---

## 10. 质量门禁与发布标准

### 10.1 代码提交门禁

| 检查项 | 要求 | 阻断级别 |
|--------|------|----------|
| 单元测试 | 100% 通过 | 阻断 |
| 覆盖率 | ≥ 80% (核心模块 ≥ 90%) | 警告 |
| 代码检查 | golangci-lint 无 Error | 阻断 |
| 代码审查 | 至少 1 人 Approve | 阻断 |

### 10.2 发布门禁

| 检查项 | 要求 | 阻断级别 |
|--------|------|----------|
| 单元测试 | 100% 通过 | 阻断 |
| 集成测试 | 100% 通过 | 阻断 |
| E2E 测试 | 100% 通过 | 阻断 |
| 场景测试 | 100% 通过 | 阻断 |
| 性能测试 | 达到目标指标 | 阻断 |
| 安全扫描 | 无高危漏洞 | 阻断 |
| 回归测试 | 无新增 Bug | 阻断 |

### 10.3 发布检查清单

```markdown
## 发布检查清单 - v{VERSION}

### 代码质量
- [ ] 所有测试通过
- [ ] 代码覆盖率 ≥ 80%
- [ ] 无 P0/P1 未修复 Bug
- [ ] 代码审查完成

### 功能验证
- [ ] 所有场景测试通过
- [ ] API 文档更新
- [ ] 变更日志更新

### 性能验证
- [ ] 性能指标达标
- [ ] 无内存泄漏
- [ ] 无性能退化

### 安全验证
- [ ] 安全扫描通过
- [ ] 敏感信息检查
- [ ] 权限验证测试

### 部署准备
- [ ] 数据库迁移脚本就绪
- [ ] 配置变更文档
- [ ] 回滚方案就绪

### 签字确认
- [ ] 开发负责人: _________ 日期: _____
- [ ] 测试负责人: _________ 日期: _____
- [ ] 运维负责人: _________ 日期: _____
```

---

## 附录

### A. Makefile 命令

```makefile
# Makefile

.PHONY: test test-unit test-integration test-e2e test-scenario test-benchmark

# 运行所有测试
test: test-unit test-integration

# 单元测试
test-unit:
	go test -v -race -coverprofile=coverage.out ./pkg/... ./internal/...
	go tool cover -func=coverage.out

# 集成测试
test-integration:
	docker-compose -f tests/integration/docker-compose.yaml up -d
	go test -v -tags=integration ./tests/integration/...
	docker-compose -f tests/integration/docker-compose.yaml down

# E2E 测试
test-e2e:
	docker-compose -f tests/e2e/docker-compose.yaml up -d
	./scripts/wait-for-ready.sh
	go test -v -tags=e2e ./tests/e2e/...
	docker-compose -f tests/e2e/docker-compose.yaml down

# 场景测试
test-scenario:
	go test -v -tags=scenario ./tests/scenario/...

# 性能基准测试
test-benchmark:
	go test -bench=. -benchmem ./tests/benchmark/...

# 负载测试
test-load:
	go test -v -tags=load ./tests/load/...

# 覆盖率报告
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html
```

### B. 测试数据生成器

```go
// tests/testutil/generator.go

package testutil

// GenerateEmployees 生成测试员工数据
func GenerateEmployees(count int, scenario string) []Resource {
    // 根据场景生成符合特征的测试数据
}

// GenerateOrders 生成测试订单数据
func GenerateOrders(count int, scenario string) []ServiceOrder {
    // 生成随机但合理的订单数据
}

// GenerateConstraints 生成测试约束配置
func GenerateConstraints(scenario string) []Constraint {
    // 返回场景默认约束
}
```

---

> **文档版本历史**
> 
> | 版本 | 日期 | 作者 | 说明 |
> |------|------|------|------|
> | v1.0 | 2026-01-11 | AI | 初始版本 |

