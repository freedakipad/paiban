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

// TestRestaurantBasicSchedule 餐饮基础排班测试
func TestRestaurantBasicSchedule(t *testing.T) {
	// 创建约束管理器
	cm := constraint.NewManager()
	builtin.RegisterDefaultConstraints(cm, map[string]interface{}{
		"max_hours_per_day":  10,
		"max_hours_per_week": 44,
	})

	// 创建排班上下文
	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-21")

	// 创建员工
	employees := []*model.Employee{
		createEmployee("张三", "服务员", []string{"收银", "点餐"}),
		createEmployee("李四", "服务员", []string{"点餐"}),
		createEmployee("王五", "厨师", []string{"烹饪"}),
		createEmployee("赵六", "服务员", []string{"收银", "点餐", "清洁"}),
	}
	ctx.SetEmployees(employees)

	// 创建班次
	shifts := []*model.Shift{
		createShift("早班", "M", "08:00", "16:00", 480, "morning"),
		createShift("晚班", "E", "16:00", "24:00", 480, "evening"),
	}
	ctx.SetShifts(shifts)

	// 创建需求 - 一周排班
	var requirements []*model.ShiftRequirement
	for day := 15; day <= 21; day++ {
		date := time.Date(2024, 1, day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		requirements = append(requirements,
			createRequirement(shifts[0].ID, date, 2, 8), // 早班需要2人
			createRequirement(shifts[1].ID, date, 1, 7), // 晚班需要1人
		)
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
	t.Logf("约束得分: %.1f", result.ConstraintResult.Score)

	// 基本验证
	if result.Statistics.TotalAssignments == 0 {
		t.Error("应该有排班分配")
	}

	if result.Statistics.FillRate < 50 {
		t.Errorf("满足率过低: %.1f%%", result.Statistics.FillRate)
	}

	// 验证每个员工的工时
	empHours := make(map[uuid.UUID]float64)
	for _, a := range result.Assignments {
		empHours[a.EmployeeID] += a.WorkingHours()
	}

	for _, emp := range employees {
		hours := empHours[emp.ID]
		t.Logf("员工 %s 工时: %.1f", emp.Name, hours)

		if hours > 44 {
			t.Errorf("员工 %s 周工时 %.1f 超过44小时限制", emp.Name, hours)
		}
	}
}

// TestRestaurantPeakHoursCoverage 餐饮高峰期覆盖测试
func TestRestaurantPeakHoursCoverage(t *testing.T) {
	// 创建约束管理器
	cm := constraint.NewManager()
	builtin.RegisterDefaultConstraints(cm, nil)

	// 添加高峰期约束
	cm.Register(builtin.NewPeakHoursCoverageConstraint(
		80,
		[]string{"11:00-13:00", "17:00-20:00"},
		3,
	))

	// 验证约束已注册
	if cm.Count() < 6 {
		t.Errorf("约束数量不足: %d", cm.Count())
	}

	t.Log("高峰期约束已注册")
}

// TestRestaurantConstraintViolation 约束违反检测测试
func TestRestaurantConstraintViolation(t *testing.T) {
	cm := constraint.NewManager()
	builtin.RegisterDefaultConstraints(cm, map[string]interface{}{
		"max_hours_per_day": 8,
	})

	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-15")

	emp := createEmployee("测试员工", "服务员", nil)
	ctx.SetEmployees([]*model.Employee{emp})

	// 创建超时分配
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(emp.ID, uuid.New(), "2024-01-15", "08:00", "18:00"), // 10小时
	})

	// 评估约束
	result := cm.Evaluate(ctx)

	if result.IsValid {
		t.Error("应该检测到约束违反（超过8小时限制）")
	}

	if len(result.HardViolations) == 0 {
		t.Error("应该有硬约束违反记录")
	}

	t.Logf("检测到 %d 个硬约束违反", len(result.HardViolations))
	for _, v := range result.HardViolations {
		t.Logf("  - %s: %s", v.ConstraintName, v.Message)
	}
}

// 辅助函数

func createEmployee(name, position string, skills []string) *model.Employee {
	return &model.Employee{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Name:      name,
		Position:  position,
		Skills:    skills,
		Status:    "active",
	}
}

func createShift(name, code, startTime, endTime string, duration int, shiftType string) *model.Shift {
	return &model.Shift{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Name:      name,
		Code:      code,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		ShiftType: shiftType,
		IsActive:  true,
	}
}

func createRequirement(shiftID uuid.UUID, date string, minEmp, priority int) *model.ShiftRequirement {
	return &model.ShiftRequirement{
		BaseModel:    model.BaseModel{ID: uuid.New()},
		ShiftID:      shiftID,
		Date:         date,
		MinEmployees: minEmp,
		MaxEmployees: minEmp * 2,
		Priority:     priority,
	}
}

func createAssignment(empID, shiftID uuid.UUID, date, start, end string) *model.Assignment {
	startTime, _ := time.Parse("2006-01-02 15:04", date+" "+start)
	endTime, _ := time.Parse("2006-01-02 15:04", date+" "+end)

	return &model.Assignment{
		BaseModel:  model.BaseModel{ID: uuid.New()},
		EmployeeID: empID,
		ShiftID:    shiftID,
		Date:       date,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     "scheduled",
	}
}
