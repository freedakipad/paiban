// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"sort"
	"time"

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

// Evaluate 评估整个排班 - 按周分割计算工时
func (c *MaxHoursPerWeekConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 获取排班周期的周列表
	weeks := c.getWeeksInRange(ctx.StartDate, ctx.EndDate)

	// 计算每个员工在每周的工时
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		// 按周分组计算工时
		hoursByWeek := make(map[string]float64)
		for _, a := range assignments {
			weekStart := c.getWeekStart(a.Date)
			hoursByWeek[weekStart] += a.WorkingHours()
		}

		// 检查每周是否超时
		for _, weekStart := range weeks {
			hours := hoursByWeek[weekStart]
			if hours > float64(c.maxHours) {
				isValid = false
				penalty := c.Weight() * int(hours-float64(c.maxHours))
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           weekStart,
					Message:        fmt.Sprintf("员工 %s 在周 %s 工作 %.1f 小时，超过限制 %d 小时", emp.Name, weekStart, hours, c.maxHours),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// getWeekStart 获取日期所在周的开始日期（周日）
func (c *MaxHoursPerWeekConstraint) getWeekStart(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	// 计算到周日的天数偏移
	weekday := int(t.Weekday())
	weekStart := t.AddDate(0, 0, -weekday)
	return weekStart.Format("2006-01-02")
}

// getWeeksInRange 获取日期范围内的所有周的起始日期
func (c *MaxHoursPerWeekConstraint) getWeeksInRange(startDate, endDate string) []string {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil
	}

	weeksMap := make(map[string]bool)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		weekStart := c.getWeekStart(d.Format("2006-01-02"))
		weeksMap[weekStart] = true
	}

	weeks := make([]string, 0, len(weeksMap))
	for w := range weeksMap {
		weeks = append(weeks, w)
	}
	sort.Strings(weeks)
	return weeks
}

// EvaluateAssignment 评估单个分配 - 计算该分配所在周的工时
func (c *MaxHoursPerWeekConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 计算该员工在该分配所在周的已有工时
	weekStart := c.getWeekStart(a.Date)
	weekEnd := c.getWeekEnd(weekStart)

	currentHours := ctx.GetEmployeeHoursInRange(a.EmployeeID, weekStart, weekEnd)
	newHours := a.WorkingHours()
	totalHours := currentHours + newHours

	if totalHours > float64(c.maxHours) {
		penalty := c.Weight() * int(totalHours-float64(c.maxHours))
		return false, penalty
	}

	return true, 0
}

// getWeekEnd 获取周结束日期（周六）
func (c *MaxHoursPerWeekConstraint) getWeekEnd(weekStartStr string) string {
	t, err := time.Parse("2006-01-02", weekStartStr)
	if err != nil {
		return weekStartStr
	}
	weekEnd := t.AddDate(0, 0, 6)
	return weekEnd.Format("2006-01-02")
}

// MaxHoursPerPeriodConstraint 排班周期最大工时约束（支持月度工时）
// 适用于按月度或其他长周期计算工时的场景
type MaxHoursPerPeriodConstraint struct {
	*BaseConstraint
	maxHours int
}

// NewMaxHoursPerPeriodConstraint 创建排班周期最大工时约束
func NewMaxHoursPerPeriodConstraint(maxHours int) *MaxHoursPerPeriodConstraint {
	return &MaxHoursPerPeriodConstraint{
		BaseConstraint: NewBaseConstraint(
			"排班周期最大工时",
			constraint.Type("max_hours_per_period"),
			constraint.CategoryHard,
			100,
		),
		maxHours: maxHours,
	}
}

// Evaluate 评估整个排班 - 计算整个排班周期内的总工时
func (c *MaxHoursPerPeriodConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 计算每个员工在整个排班周期内的总工时
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
				Message:        fmt.Sprintf("员工 %s 在排班周期内工作 %.1f 小时，超过限制 %d 小时", emp.Name, totalHours, c.maxHours),
				Severity:       "error",
				Penalty:        penalty,
			})
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配 - 计算整个排班周期的工时
func (c *MaxHoursPerPeriodConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 计算该员工在整个排班周期内的已有工时
	currentHours := ctx.GetEmployeeHoursInRange(a.EmployeeID, ctx.StartDate, ctx.EndDate)
	newHours := a.WorkingHours()
	totalHours := currentHours + newHours

	if totalHours > float64(c.maxHours) {
		penalty := c.Weight() * int(totalHours-float64(c.maxHours))
		return false, penalty
	}

	return true, 0
}
