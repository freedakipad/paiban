// Package scenario 提供场景测试
package scenario

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
	"github.com/paiban/paiban/pkg/scheduler/constraint/builtin"
	"github.com/paiban/paiban/pkg/scheduler/solver"
)

// TestFactoryThreeShiftSchedule 工厂三班倒排班测试
func TestFactoryThreeShiftSchedule(t *testing.T) {
	// 创建约束管理器
	cm := constraint.NewManager()
	builtin.RegisterDefaultConstraints(cm, map[string]interface{}{
		"max_hours_per_day":  8,
		"max_hours_per_week": 44,
	})

	// 添加工厂特有约束
	cm.Register(builtin.NewMaxConsecutiveNightsConstraint(4))
	cm.Register(builtin.NewShiftRotationPatternConstraint(100, "三班倒", 7))

	// 创建排班上下文
	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-21")

	// 创建员工（产线工人）
	employees := []*model.Employee{
		createFactoryEmployee("工人A", "操作工", []string{"CNC操作", "质检"}, "team_1"),
		createFactoryEmployee("工人B", "操作工", []string{"CNC操作"}, "team_1"),
		createFactoryEmployee("工人C", "技术员", []string{"维修", "调试"}, "team_1"),
		createFactoryEmployee("工人D", "操作工", []string{"CNC操作", "质检"}, "team_2"),
		createFactoryEmployee("工人E", "操作工", []string{"CNC操作"}, "team_2"),
		createFactoryEmployee("工人F", "技术员", []string{"维修"}, "team_2"),
	}
	ctx.SetEmployees(employees)

	// 创建三班倒班次
	shifts := []*model.Shift{
		createShift("早班", "A", "08:00", "16:00", 480, "morning"),
		createShift("中班", "B", "16:00", "24:00", 480, "afternoon"),
		createShift("夜班", "C", "00:00", "08:00", 480, "night"),
	}
	ctx.SetShifts(shifts)

	// 创建需求 - 每班需要2人
	var requirements []*model.ShiftRequirement
	for day := 15; day <= 21; day++ {
		date := time.Date(2024, 1, day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		for _, shift := range shifts {
			requirements = append(requirements, createRequirement(shift.ID, date, 2, 8))
		}
	}
	ctx.Requirements = requirements

	// 执行排班
	s := solver.NewGreedySolver(cm)
	result, err := s.Solve(context.Background(), ctx)

	if err != nil {
		t.Fatalf("排班执行失败: %v", err)
	}

	// 验证结果
	t.Logf("排班成功: %v", result.Success)
	t.Logf("总分配数: %d", result.Statistics.TotalAssignments)
	t.Logf("满足率: %.1f%%", result.Statistics.FillRate)
	t.Logf("执行时间: %v", result.Duration)

	// 统计每个员工的夜班数
	nightShiftCount := make(map[uuid.UUID]int)
	for _, a := range result.Assignments {
		shift := ctx.GetShift(a.ShiftID)
		if shift != nil && shift.IsNightShift() {
			nightShiftCount[a.EmployeeID]++
		}
	}

	for _, emp := range employees {
		count := nightShiftCount[emp.ID]
		t.Logf("员工 %s 夜班数: %d", emp.Name, count)

		// 验证夜班不超过4个
		if count > 4 {
			t.Errorf("员工 %s 连续夜班 %d 个，超过限制4个", emp.Name, count)
		}
	}
}

// TestFactoryMaxConsecutiveNights 最大连续夜班测试
func TestFactoryMaxConsecutiveNights(t *testing.T) {
	constraint := builtin.NewMaxConsecutiveNightsConstraint(4)

	if constraint.Name() != "最大连续夜班" {
		t.Errorf("约束名称错误: %s", constraint.Name())
	}

	if constraint.Category() != "hard" {
		t.Errorf("约束类别错误: %s", constraint.Category())
	}

	t.Log("最大连续夜班约束创建成功")
}

// TestFactoryShiftRotation 倒班模式测试
func TestFactoryShiftRotation(t *testing.T) {
	cm := constraint.NewManager()
	cm.Register(builtin.NewShiftRotationPatternConstraint(100, "三班倒", 7))

	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-17")

	emp := createFactoryEmployee("测试工人", "操作工", nil, "team_1")
	ctx.SetEmployees([]*model.Employee{emp})

	nightShift := createShift("夜班", "C", "00:00", "08:00", 480, "night")
	morningShift := createShift("早班", "A", "08:00", "16:00", 480, "morning")
	ctx.SetShifts([]*model.Shift{nightShift, morningShift})

	// 创建违规场景：夜班后次日早班
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(emp.ID, nightShift.ID, "2024-01-15", "00:00", "08:00"),   // 夜班
		createAssignment(emp.ID, morningShift.ID, "2024-01-16", "08:00", "16:00"), // 次日早班
	})

	// 评估约束
	result := cm.Evaluate(ctx)

	if result.IsValid {
		t.Error("应该检测到倒班模式违规（夜班后次日早班）")
	}

	t.Logf("约束评估结果: 有效=%v, 得分=%.1f", result.IsValid, result.Score)
}

// TestFactoryProductionLineCoverage 产线覆盖测试
func TestFactoryProductionLineCoverage(t *testing.T) {
	requirements := map[string]int{
		"操作工": 2,
		"技术员": 1,
	}

	constraint := builtin.NewProductionLineCoverageConstraint(100, requirements)

	if constraint.Name() != "产线覆盖" {
		t.Errorf("约束名称错误: %s", constraint.Name())
	}

	t.Log("产线覆盖约束创建成功")
}

// 辅助函数

func createFactoryEmployee(name, position string, skills []string, _ string) *model.Employee {
	return &model.Employee{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Name:      name,
		Position:  position,
		Skills:    skills,
		Status:    "active",
		// 班组信息可以通过扩展字段存储
	}
}
