// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"time"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// MaxShiftsPerMonthConstraint 每月最大班次数约束
// 支持为每个月设置不同的限制
type MaxShiftsPerMonthConstraint struct {
	*BaseConstraint
	defaultMaxShifts  int            // 默认最大班次数
	monthlyMaxShifts  map[string]int // 每月单独设置的最大班次数 (key: YYYY-MM)
}

// NewMaxShiftsPerMonthConstraint 创建每月最大班次数约束
// maxShifts: 默认的最大班次数
// monthlyLimits: 可选，每月单独设置的限制 (key: YYYY-MM, value: 最大班次数)
func NewMaxShiftsPerMonthConstraint(maxShifts int, monthlyLimits ...map[string]int) *MaxShiftsPerMonthConstraint {
	c := &MaxShiftsPerMonthConstraint{
		BaseConstraint: NewBaseConstraint(
			"每月最大班次数",
			constraint.Type("max_shifts_per_month"),
			constraint.CategoryHard,
			100, // 硬约束权重
		),
		defaultMaxShifts: maxShifts,
		monthlyMaxShifts: make(map[string]int),
	}
	
	// 如果传入了每月限制，使用它
	if len(monthlyLimits) > 0 && monthlyLimits[0] != nil {
		c.monthlyMaxShifts = monthlyLimits[0]
	}
	
	return c
}

// getMaxShiftsForMonth 获取指定月份的最大班次数
func (c *MaxShiftsPerMonthConstraint) getMaxShiftsForMonth(month string) int {
	if limit, ok := c.monthlyMaxShifts[month]; ok {
		return limit
	}
	return c.defaultMaxShifts
}

// Evaluate 评估整个排班
func (c *MaxShiftsPerMonthConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 获取排班周期内的所有月份
	months := c.getMonthsInRange(ctx.StartDate, ctx.EndDate)

	// 计算每个员工在每月的班次数
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		// 按月分组计算当前上下文中的班次数
		shiftsByMonth := make(map[string]int)
		for _, a := range assignments {
			month := c.getMonth(a.Date)
			shiftsByMonth[month]++
		}

		// 检查每月是否超出限制（包含前端传入的历史班次）
		for _, month := range months {
			// 获取该月的最大班次数限制
			maxShifts := c.getMaxShiftsForMonth(month)

			// 当前排班中该月的班次数
			contextShifts := shiftsByMonth[month]

			// 加上前端传入的该月历史班次数
			existingShifts := 0
			if emp.MonthlyShiftsCounts != nil {
				existingShifts = emp.MonthlyShiftsCounts[month]
			}

			totalShifts := existingShifts + contextShifts

			if totalShifts > maxShifts {
				isValid = false
				penalty := c.Weight() * (totalShifts - maxShifts)
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           month,
					Message:        fmt.Sprintf("员工 %s 在 %s 月有 %d 个班次（历史%d+当前%d），超过限制 %d 个", emp.Name, month, totalShifts, existingShifts, contextShifts, maxShifts),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MaxShiftsPerMonthConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 获取员工
	emp := ctx.GetEmployee(a.EmployeeID)
	if emp == nil {
		return true, 0
	}

	// 获取分配日期所在的月份
	month := c.getMonth(a.Date)

	// 获取该月的最大班次数限制
	maxShifts := c.getMaxShiftsForMonth(month)

	// 获取员工该月已有班次数（前端传入的历史班次）
	existingShifts := 0
	if emp.MonthlyShiftsCounts != nil {
		existingShifts = emp.MonthlyShiftsCounts[month]
	}

	// 计算该员工在当前排班上下文中该月的已有班次数
	monthStart, monthEnd := c.getMonthRange(month)

	contextShifts := 0
	assignments := ctx.GetEmployeeAssignments(a.EmployeeID)
	for _, existing := range assignments {
		if existing.Date >= monthStart && existing.Date <= monthEnd {
			contextShifts++
		}
	}

	// 总班次 = 前端传入的该月历史班次 + 当前上下文中该月的班次 + 新分配的1个班次
	totalShifts := existingShifts + contextShifts + 1

	if totalShifts > maxShifts {
		penalty := c.Weight() * (totalShifts - maxShifts)
		return false, penalty
	}

	return true, 0
}

// getMonth 获取日期所在月份（YYYY-MM格式）
func (c *MaxShiftsPerMonthConstraint) getMonth(dateStr string) string {
	if len(dateStr) >= 7 {
		return dateStr[:7]
	}
	return dateStr
}

// getMonthsInRange 获取日期范围内的所有月份
func (c *MaxShiftsPerMonthConstraint) getMonthsInRange(startDate, endDate string) []string {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil
	}

	monthsMap := make(map[string]bool)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		month := d.Format("2006-01")
		monthsMap[month] = true
	}

	months := make([]string, 0, len(monthsMap))
	for m := range monthsMap {
		months = append(months, m)
	}
	return months
}

// getMonthRange 获取月份的起止日期
func (c *MaxShiftsPerMonthConstraint) getMonthRange(month string) (string, string) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month + "-01", month + "-31"
	}

	// 月初
	monthStart := t.Format("2006-01-02")

	// 月末（下月第一天减一天）
	nextMonth := t.AddDate(0, 1, 0)
	monthEnd := nextMonth.AddDate(0, 0, -1).Format("2006-01-02")

	return monthStart, monthEnd
}
