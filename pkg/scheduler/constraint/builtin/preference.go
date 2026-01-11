// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"time"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// EmployeePreferenceConstraint 员工偏好约束（软约束）
type EmployeePreferenceConstraint struct {
	*BaseConstraint
}

// NewEmployeePreferenceConstraint 创建员工偏好约束
func NewEmployeePreferenceConstraint(weight int) *EmployeePreferenceConstraint {
	return &EmployeePreferenceConstraint{
		BaseConstraint: NewBaseConstraint(
			"员工偏好",
			constraint.TypeEmployeePreference,
			constraint.CategorySoft,
			weight,
		),
	}
}

// Evaluate 评估整个排班
func (c *EmployeePreferenceConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	for _, a := range ctx.Assignments {
		emp := ctx.GetEmployee(a.EmployeeID)
		if emp == nil || emp.Preferences == nil {
			continue
		}

		prefs := emp.Preferences
		shift := ctx.GetShift(a.ShiftID)

		// 检查班次偏好
		if len(prefs.AvoidShifts) > 0 && shift != nil {
			for _, avoidShift := range prefs.AvoidShifts {
				if shift.Code == avoidShift || shift.ShiftType == avoidShift {
					penalty := c.Weight() / 2
					totalPenalty += penalty
					violations = append(violations, constraint.ViolationDetail{
						ConstraintType: c.Type(),
						ConstraintName: c.Name(),
						EmployeeID:     emp.ID,
						Date:           a.Date,
						Message:        fmt.Sprintf("员工 %s 希望避免班次: %s", emp.Name, avoidShift),
						Severity:       "warning",
						Penalty:        penalty,
					})
				}
			}
		}

		// 检查日期偏好
		if len(prefs.AvoidDays) > 0 {
			assignDate, err := time.Parse("2006-01-02", a.Date)
			if err == nil {
				weekday := assignDate.Weekday()
				for _, avoidDay := range prefs.AvoidDays {
					if weekday == avoidDay {
						penalty := c.Weight() / 2
						totalPenalty += penalty
						violations = append(violations, constraint.ViolationDetail{
							ConstraintType: c.Type(),
							ConstraintName: c.Name(),
							EmployeeID:     emp.ID,
							Date:           a.Date,
							Message:        fmt.Sprintf("员工 %s 希望避免在 %s 工作", emp.Name, weekday.String()),
							Severity:       "warning",
							Penalty:        penalty,
						})
					}
				}
			}
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *EmployeePreferenceConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	emp := ctx.GetEmployee(a.EmployeeID)
	if emp == nil || emp.Preferences == nil {
		return true, 0
	}

	prefs := emp.Preferences
	shift := ctx.GetShift(a.ShiftID)
	penalty := 0

	// 检查是否为避免的班次
	if len(prefs.AvoidShifts) > 0 && shift != nil {
		for _, avoidShift := range prefs.AvoidShifts {
			if shift.Code == avoidShift || shift.ShiftType == avoidShift {
				penalty += c.Weight() / 2
			}
		}
	}

	// 检查是否为避免的日期
	if len(prefs.AvoidDays) > 0 {
		assignDate, err := time.Parse("2006-01-02", a.Date)
		if err == nil {
			weekday := assignDate.Weekday()
			for _, avoidDay := range prefs.AvoidDays {
				if weekday == avoidDay {
					penalty += c.Weight() / 2
				}
			}
		}
	}

	// 检查偏好的班次（给予负惩罚/奖励）
	if len(prefs.PreferredShifts) > 0 && shift != nil {
		for _, prefShift := range prefs.PreferredShifts {
			if shift.Code == prefShift || shift.ShiftType == prefShift {
				penalty -= c.Weight() / 4 // 奖励
			}
		}
	}

	return true, penalty
}

// MinimizeOvertimeConstraint 最小化加班约束（软约束）
type MinimizeOvertimeConstraint struct {
	*BaseConstraint
	standardHoursPerWeek int
}

// NewMinimizeOvertimeConstraint 创建最小化加班约束
func NewMinimizeOvertimeConstraint(weight int, standardHours int) *MinimizeOvertimeConstraint {
	return &MinimizeOvertimeConstraint{
		BaseConstraint: NewBaseConstraint(
			"最小化加班",
			constraint.TypeMinimizeOvertime,
			constraint.CategorySoft,
			weight,
		),
		standardHoursPerWeek: standardHours,
	}
}

// Evaluate 评估整个排班
func (c *MinimizeOvertimeConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		var totalHours float64
		for _, a := range assignments {
			totalHours += a.WorkingHours()
		}

		overtime := totalHours - float64(c.standardHoursPerWeek)
		if overtime > 0 {
			penalty := int(overtime * float64(c.Weight()) / 10)
			totalPenalty += penalty
			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message:        fmt.Sprintf("员工 %s 加班 %.1f 小时", emp.Name, overtime),
				Severity:       "warning",
				Penalty:        penalty,
			})
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MinimizeOvertimeConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	currentHours := ctx.GetEmployeeHoursInRange(a.EmployeeID, ctx.StartDate, ctx.EndDate)
	newHours := a.WorkingHours()
	totalHours := currentHours + newHours

	if totalHours > float64(c.standardHoursPerWeek) {
		overtime := totalHours - float64(c.standardHoursPerWeek)
		penalty := int(overtime * float64(c.Weight()) / 10)
		return true, penalty
	}

	return true, 0
}
