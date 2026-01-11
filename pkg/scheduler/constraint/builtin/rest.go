// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"sort"
	"time"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// MinRestBetweenShiftsConstraint 班次间最小休息时间约束
type MinRestBetweenShiftsConstraint struct {
	*BaseConstraint
	minHours int
}

// NewMinRestBetweenShiftsConstraint 创建班次间最小休息约束
func NewMinRestBetweenShiftsConstraint(minHours int) *MinRestBetweenShiftsConstraint {
	return &MinRestBetweenShiftsConstraint{
		BaseConstraint: NewBaseConstraint(
			"班次间最小休息",
			constraint.TypeMinRestBetweenShifts,
			constraint.CategoryHard,
			100,
		),
		minHours: minHours,
	}
}

// Evaluate 评估整个排班
func (c *MinRestBetweenShiftsConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		if len(assignments) < 2 {
			continue
		}

		// 按时间排序
		sorted := make([]*model.Assignment, len(assignments))
		copy(sorted, assignments)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].EndTime.Before(sorted[j].EndTime)
		})

		// 检查相邻班次间隔
		for i := 0; i < len(sorted)-1; i++ {
			restHours := sorted[i+1].StartTime.Sub(sorted[i].EndTime).Hours()

			if restHours < float64(c.minHours) {
				isValid = false
				penalty := c.Weight() * int(float64(c.minHours)-restHours)
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           sorted[i+1].Date,
					Message: fmt.Sprintf(
						"员工 %s 班次间隔仅 %.1f 小时，少于要求的 %d 小时",
						emp.Name, restHours, c.minHours,
					),
					Severity: "error",
					Penalty:  penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MinRestBetweenShiftsConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	assignments := ctx.GetEmployeeAssignments(a.EmployeeID)

	for _, existing := range assignments {
		if existing.ID == a.ID {
			continue
		}

		// 检查与现有班次的间隔
		var restHours float64
		if a.StartTime.After(existing.EndTime) {
			restHours = a.StartTime.Sub(existing.EndTime).Hours()
		} else if existing.StartTime.After(a.EndTime) {
			restHours = existing.StartTime.Sub(a.EndTime).Hours()
		} else {
			// 班次重叠
			return false, c.Weight() * c.minHours
		}

		if restHours < float64(c.minHours) {
			penalty := c.Weight() * int(float64(c.minHours)-restHours)
			return false, penalty
		}
	}

	return true, 0
}

// MaxConsecutiveDaysConstraint 最大连续工作天数约束
type MaxConsecutiveDaysConstraint struct {
	*BaseConstraint
	maxDays int
}

// NewMaxConsecutiveDaysConstraint 创建最大连续工作天数约束
func NewMaxConsecutiveDaysConstraint(maxDays int) *MaxConsecutiveDaysConstraint {
	return &MaxConsecutiveDaysConstraint{
		BaseConstraint: NewBaseConstraint(
			"最大连续工作天数",
			constraint.TypeMaxConsecutiveDays,
			constraint.CategoryHard,
			100,
		),
		maxDays: maxDays,
	}
}

// Evaluate 评估整个排班
func (c *MaxConsecutiveDaysConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		if len(assignments) == 0 {
			continue
		}

		// 获取所有工作日期
		workDates := make(map[string]bool)
		for _, a := range assignments {
			workDates[a.Date] = true
		}

		// 将日期排序
		dates := make([]string, 0, len(workDates))
		for d := range workDates {
			dates = append(dates, d)
		}
		sort.Strings(dates)

		// 检查连续天数
		consecutive := 1
		maxConsecutive := 1
		for i := 1; i < len(dates); i++ {
			// 简化实现：假设日期格式正确且连续
			// 实际应该计算日期差
			if isConsecutiveDate(dates[i-1], dates[i]) {
				consecutive++
				if consecutive > maxConsecutive {
					maxConsecutive = consecutive
				}
			} else {
				consecutive = 1
			}
		}

		if maxConsecutive > c.maxDays {
			isValid = false
			penalty := c.Weight() * (maxConsecutive - c.maxDays)
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message: fmt.Sprintf(
					"员工 %s 连续工作 %d 天，超过限制 %d 天",
					emp.Name, maxConsecutive, c.maxDays,
				),
				Severity: "error",
				Penalty:  penalty,
			})
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MaxConsecutiveDaysConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 计算加上新分配后的连续天数
	consecutiveDays := ctx.GetEmployeeConsecutiveDays(a.EmployeeID, a.Date) + 1

	if consecutiveDays > c.maxDays {
		penalty := c.Weight() * (consecutiveDays - c.maxDays)
		return false, penalty
	}

	return true, 0
}

// isConsecutiveDate 检查两个日期是否连续
// 支持同月和跨月场景
func isConsecutiveDate(date1, date2 string) bool {
	if len(date1) != 10 || len(date2) != 10 {
		return false
	}

	// 解析日期
	t1, err1 := time.Parse("2006-01-02", date1)
	t2, err2 := time.Parse("2006-01-02", date2)
	if err1 != nil || err2 != nil {
		return false
	}

	// 计算日期差值
	diff := t2.Sub(t1)
	return diff == 24*time.Hour
}
