// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// MaxHoursPerDayConstraint 每日最大工时约束
type MaxHoursPerDayConstraint struct {
	*BaseConstraint
	maxHours int
}

// NewMaxHoursPerDayConstraint 创建每日最大工时约束
func NewMaxHoursPerDayConstraint(maxHours int) *MaxHoursPerDayConstraint {
	return &MaxHoursPerDayConstraint{
		BaseConstraint: NewBaseConstraint(
			"每日最大工时",
			constraint.TypeMaxHoursPerDay,
			constraint.CategoryHard,
			100, // 硬约束权重
		),
		maxHours: maxHours,
	}
}

// Evaluate 评估整个排班
func (c *MaxHoursPerDayConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 遍历所有员工
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		// 按日期分组计算工时
		hoursByDate := make(map[string]float64)
		for _, a := range assignments {
			hoursByDate[a.Date] += a.WorkingHours()
		}

		// 检查每天是否超时
		for date, hours := range hoursByDate {
			if hours > float64(c.maxHours) {
				isValid = false
				penalty := c.Weight() * int(hours-float64(c.maxHours))
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           date,
					Message:        fmt.Sprintf("员工 %s 在 %s 工作 %.1f 小时，超过限制 %d 小时", emp.Name, date, hours, c.maxHours),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MaxHoursPerDayConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 计算该员工当天已有工时 + 新分配工时
	currentHours := ctx.GetEmployeeHoursOnDate(a.EmployeeID, a.Date)
	newHours := a.WorkingHours()
	totalHours := currentHours + newHours

	if totalHours > float64(c.maxHours) {
		penalty := c.Weight() * int(totalHours-float64(c.maxHours))
		return false, penalty
	}

	return true, 0
}

// MaxHoursPerWeekConstraint 每周最大工时约束
type MaxHoursPerWeekConstraint struct {
	*BaseConstraint
	maxHours int
}

// NewMaxHoursPerWeekConstraint 创建每周最大工时约束
func NewMaxHoursPerWeekConstraint(maxHours int) *MaxHoursPerWeekConstraint {
	return &MaxHoursPerWeekConstraint{
		BaseConstraint: NewBaseConstraint(
			"每周最大工时",
			constraint.TypeMaxHoursPerWeek,
			constraint.CategoryHard,
			100,
		),
		maxHours: maxHours,
	}
}

// Evaluate 评估整个排班
func (c *MaxHoursPerWeekConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 计算每个员工的总工时
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		var totalHours float64
		for _, a := range assignments {
			totalHours += a.WorkingHours()
		}

		if totalHours > float64(c.maxHours) {
			isValid = false
			penalty := c.Weight() * int(totalHours-float64(c.maxHours))
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message:        fmt.Sprintf("员工 %s 周工作 %.1f 小时，超过限制 %d 小时", emp.Name, totalHours, c.maxHours),
				Severity:       "error",
				Penalty:        penalty,
			})
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MaxHoursPerWeekConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 计算该员工本周已有工时
	currentHours := ctx.GetEmployeeHoursInRange(a.EmployeeID, ctx.StartDate, ctx.EndDate)
	newHours := a.WorkingHours()
	totalHours := currentHours + newHours

	if totalHours > float64(c.maxHours) {
		penalty := c.Weight() * int(totalHours-float64(c.maxHours))
		return false, penalty
	}

	return true, 0
}
