package builtin

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

func TestMaxHoursPerDayConstraint_Evaluate(t *testing.T) {
	tests := []struct {
		name        string
		maxHours    int
		assignments []*model.Assignment
		wantValid   bool
		wantPenalty int
	}{
		{
			name:        "无分配，应通过",
			maxHours:    10,
			assignments: nil,
			wantValid:   true,
			wantPenalty: 0,
		},
		{
			name:     "工时未超限，应通过",
			maxHours: 10,
			assignments: []*model.Assignment{
				createAssignment("2024-01-15", 8), // 8小时
			},
			wantValid:   true,
			wantPenalty: 0,
		},
		{
			name:     "工时超限，应失败",
			maxHours: 8,
			assignments: []*model.Assignment{
				createAssignment("2024-01-15", 10), // 10小时，超过8小时限制
			},
			wantValid:   false,
			wantPenalty: 200, // 100 * 2
		},
		{
			name:     "同一天多个班次超限",
			maxHours: 10,
			assignments: []*model.Assignment{
				createAssignmentWithTime("2024-01-15", "08:00", "14:00"), // 6小时
				createAssignmentWithTime("2024-01-15", "16:00", "22:00"), // 6小时，总12小时
			},
			wantValid:   false,
			wantPenalty: 200, // 100 * 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMaxHoursPerDayConstraint(tt.maxHours)
			ctx := createTestContext(tt.assignments)

			valid, penalty, _ := c.Evaluate(ctx)

			if valid != tt.wantValid {
				t.Errorf("Evaluate() valid = %v, want %v", valid, tt.wantValid)
			}
			if penalty != tt.wantPenalty {
				t.Errorf("Evaluate() penalty = %v, want %v", penalty, tt.wantPenalty)
			}
		})
	}
}

func TestMaxHoursPerDayConstraint_EvaluateAssignment(t *testing.T) {
	c := NewMaxHoursPerDayConstraint(10)

	// 创建上下文，已有6小时的分配
	ctx := createTestContext([]*model.Assignment{
		createAssignmentWithTime("2024-01-15", "08:00", "14:00"), // 6小时
	})

	// 获取测试员工ID
	empID := ctx.Employees[0].ID

	// 测试新增4小时（总10小时）应该通过
	newAssignment := createAssignmentWithTime("2024-01-15", "16:00", "20:00") // 4小时
	newAssignment.EmployeeID = empID
	valid, penalty := c.EvaluateAssignment(ctx, newAssignment)
	if !valid || penalty != 0 {
		t.Errorf("10小时未超限应通过，got valid=%v, penalty=%d", valid, penalty)
	}

	// 测试新增6小时（总12小时）应该失败
	overAssignment := createAssignmentWithTime("2024-01-15", "16:00", "22:00") // 6小时
	overAssignment.EmployeeID = empID
	valid, penalty = c.EvaluateAssignment(ctx, overAssignment)
	if valid || penalty == 0 {
		t.Errorf("12小时超限应失败，got valid=%v, penalty=%d", valid, penalty)
	}
}

func TestMaxHoursPerWeekConstraint_Evaluate(t *testing.T) {
	c := NewMaxHoursPerWeekConstraint(44)

	// 创建一周的分配，每天8小时，共6天 = 48小时
	var assignments []*model.Assignment
	for i := 15; i <= 20; i++ {
		date := time.Date(2024, 1, i, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		assignments = append(assignments, createAssignmentOnDate(date, 8))
	}

	ctx := createTestContext(assignments)
	valid, penalty, violations := c.Evaluate(ctx)

	if valid {
		t.Error("48小时超过44小时限制应该失败")
	}
	if penalty == 0 {
		t.Error("应该有惩罚值")
	}
	if len(violations) == 0 {
		t.Error("应该有违反详情")
	}
}

// 辅助函数

func createTestContext(assignments []*model.Assignment) *constraint.Context {
	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-21")

	// 创建测试员工
	empID := uuid.New()
	emp := &model.Employee{
		BaseModel: model.BaseModel{ID: empID},
		Name:      "测试员工",
		Status:    "active",
	}
	ctx.SetEmployees([]*model.Employee{emp})

	// 设置员工ID到所有分配
	for _, a := range assignments {
		a.EmployeeID = empID
	}

	ctx.SetAssignments(assignments)
	return ctx
}

func createAssignment(date string, hours int) *model.Assignment {
	return createAssignmentOnDate(date, hours)
}

func createAssignmentOnDate(date string, hours int) *model.Assignment {
	startTime, _ := time.Parse("2006-01-02 15:04", date+" 09:00")
	endTime := startTime.Add(time.Duration(hours) * time.Hour)

	return &model.Assignment{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "scheduled",
	}
}

func createAssignmentWithTime(date, start, end string) *model.Assignment {
	startTime, _ := time.Parse("2006-01-02 15:04", date+" "+start)
	endTime, _ := time.Parse("2006-01-02 15:04", date+" "+end)

	return &model.Assignment{
		BaseModel: model.BaseModel{ID: uuid.New()},
		Date:      date,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "scheduled",
	}
}
