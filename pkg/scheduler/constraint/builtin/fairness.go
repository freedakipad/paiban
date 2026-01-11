// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"math"
	"time"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// WorkloadFairnessConstraint 工作量公平性约束（增强版）
// 考虑多个维度的公平性
type WorkloadFairnessConstraint struct {
	*BaseConstraint
	tolerancePercent float64 // 允许的偏差百分比
	considerWeekend  bool    // 是否考虑周末分配公平
	considerNight    bool    // 是否考虑夜班分配公平
}

// NewWorkloadFairnessConstraint 创建工作量公平性约束
func NewWorkloadFairnessConstraint(weight int, tolerance float64) *WorkloadFairnessConstraint {
	return &WorkloadFairnessConstraint{
		BaseConstraint: NewBaseConstraint(
			"工作量公平性",
			constraint.Type("workload_fairness"),
			constraint.CategorySoft,
			weight,
		),
		tolerancePercent: tolerance,
		considerWeekend:  true,
		considerNight:    true,
	}
}

// Evaluate 评估整个排班
func (c *WorkloadFairnessConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	if len(ctx.Employees) < 2 {
		return true, 0, nil
	}

	// 计算工时公平性
	hoursViolations, hoursPenalty := c.evaluateHoursFairness(ctx)
	violations = append(violations, hoursViolations...)
	totalPenalty += hoursPenalty

	// 计算周末分配公平性
	if c.considerWeekend {
		weekendViolations, weekendPenalty := c.evaluateWeekendFairness(ctx)
		violations = append(violations, weekendViolations...)
		totalPenalty += weekendPenalty
	}

	// 计算夜班分配公平性
	if c.considerNight {
		nightViolations, nightPenalty := c.evaluateNightFairness(ctx)
		violations = append(violations, nightViolations...)
		totalPenalty += nightPenalty
	}

	return true, totalPenalty, violations
}

// evaluateHoursFairness 评估工时公平性
func (c *WorkloadFairnessConstraint) evaluateHoursFairness(ctx *constraint.Context) ([]constraint.ViolationDetail, int) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	// 计算每人工时
	hours := make([]float64, len(ctx.Employees))
	for i, emp := range ctx.Employees {
		hours[i] = ctx.GetEmployeeHoursInRange(emp.ID, ctx.StartDate, ctx.EndDate)
	}

	// 计算统计量
	avg, stdDev := calculateStats(hours)
	tolerance := avg * c.tolerancePercent / 100

	// 检查偏差
	for i, emp := range ctx.Employees {
		deviation := hours[i] - avg
		if math.Abs(deviation) > tolerance {
			penalty := int(math.Abs(deviation) * float64(c.Weight()) / (avg + 1))
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message: fmt.Sprintf(
					"员工 %s 工时 %.1f 小时，偏离平均 %.1f 小时 (平均: %.1f, 标准差: %.1f)",
					emp.Name, hours[i], deviation, avg, stdDev,
				),
				Severity: "warning",
				Penalty:  penalty,
			})
		}
	}

	return violations, totalPenalty
}

// evaluateWeekendFairness 评估周末分配公平性
func (c *WorkloadFairnessConstraint) evaluateWeekendFairness(ctx *constraint.Context) ([]constraint.ViolationDetail, int) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	// 统计每人周末工作天数
	weekendDays := make(map[string]int)
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		count := 0
		for _, a := range assignments {
			if isWeekend(a.Date) {
				count++
			}
		}
		weekendDays[emp.ID.String()] = count
	}

	// 计算平均值
	var total int
	for _, count := range weekendDays {
		total += count
	}
	avg := float64(total) / float64(len(ctx.Employees))

	// 检查偏差
	for _, emp := range ctx.Employees {
		count := weekendDays[emp.ID.String()]
		deviation := float64(count) - avg

		if math.Abs(deviation) > 1 { // 允许1天偏差
			penalty := int(math.Abs(deviation)) * c.Weight() / 4
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message: fmt.Sprintf(
					"员工 %s 周末工作 %d 天，偏离平均 %.1f 天",
					emp.Name, count, deviation,
				),
				Severity: "warning",
				Penalty:  penalty,
			})
		}
	}

	return violations, totalPenalty
}

// evaluateNightFairness 评估夜班分配公平性
func (c *WorkloadFairnessConstraint) evaluateNightFairness(ctx *constraint.Context) ([]constraint.ViolationDetail, int) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	// 统计每人夜班数
	nightShifts := make(map[string]int)
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		count := 0
		for _, a := range assignments {
			shift := ctx.GetShift(a.ShiftID)
			if shift != nil && shift.IsNightShift() {
				count++
			}
		}
		nightShifts[emp.ID.String()] = count
	}

	// 计算平均值
	var total int
	for _, count := range nightShifts {
		total += count
	}
	avg := float64(total) / float64(len(ctx.Employees))

	// 检查偏差
	for _, emp := range ctx.Employees {
		count := nightShifts[emp.ID.String()]
		deviation := float64(count) - avg

		if math.Abs(deviation) > 1 { // 允许1班偏差
			penalty := int(math.Abs(deviation)) * c.Weight() / 4
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message: fmt.Sprintf(
					"员工 %s 夜班 %d 次，偏离平均 %.1f 次",
					emp.Name, count, deviation,
				),
				Severity: "warning",
				Penalty:  penalty,
			})
		}
	}

	return violations, totalPenalty
}

// EvaluateAssignment 评估单个分配
func (c *WorkloadFairnessConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 软约束，总是允许
	return true, 0
}

// SeniorityBalanceConstraint 工龄均衡约束
// 确保资深员工和新员工的工作量分配合理
type SeniorityBalanceConstraint struct {
	*BaseConstraint
	seniorThresholdMonths int // 资深员工阈值（月）
}

// NewSeniorityBalanceConstraint 创建工龄均衡约束
func NewSeniorityBalanceConstraint(weight int, threshold int) *SeniorityBalanceConstraint {
	return &SeniorityBalanceConstraint{
		BaseConstraint: NewBaseConstraint(
			"工龄均衡",
			constraint.Type("seniority_balance"),
			constraint.CategorySoft,
			weight,
		),
		seniorThresholdMonths: threshold,
	}
}

// Evaluate 评估整个排班
func (c *SeniorityBalanceConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	// 简化实现，可根据员工入职日期计算
	return true, 0, nil
}

// EvaluateAssignment 评估单个分配
func (c *SeniorityBalanceConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// ShiftDistributionConstraint 班次分配均衡约束
// 确保各类班次在员工间均匀分配
type ShiftDistributionConstraint struct {
	*BaseConstraint
}

// NewShiftDistributionConstraint 创建班次分配均衡约束
func NewShiftDistributionConstraint(weight int) *ShiftDistributionConstraint {
	return &ShiftDistributionConstraint{
		BaseConstraint: NewBaseConstraint(
			"班次分配均衡",
			constraint.Type("shift_distribution"),
			constraint.CategorySoft,
			weight,
		),
	}
}

// Evaluate 评估整个排班
func (c *ShiftDistributionConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	if len(ctx.Employees) < 2 || len(ctx.Shifts) < 2 {
		return true, 0, nil
	}

	// 统计每人每种班次的数量
	shiftCounts := make(map[string]map[string]int) // empID -> shiftType -> count
	for _, emp := range ctx.Employees {
		shiftCounts[emp.ID.String()] = make(map[string]int)
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		for _, a := range assignments {
			shift := ctx.GetShift(a.ShiftID)
			if shift != nil {
				shiftCounts[emp.ID.String()][shift.ShiftType]++
			}
		}
	}

	// 计算每种班次的平均分配
	shiftTotals := make(map[string]int)
	for _, counts := range shiftCounts {
		for shiftType, count := range counts {
			shiftTotals[shiftType] += count
		}
	}

	shiftAvg := make(map[string]float64)
	for shiftType, total := range shiftTotals {
		shiftAvg[shiftType] = float64(total) / float64(len(ctx.Employees))
	}

	// 检查偏差
	for _, emp := range ctx.Employees {
		counts := shiftCounts[emp.ID.String()]
		for shiftType, count := range counts {
			avg := shiftAvg[shiftType]
			deviation := float64(count) - avg

			if math.Abs(deviation) > 2 { // 允许2班偏差
				penalty := int(math.Abs(deviation)) * c.Weight() / 5
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Message: fmt.Sprintf(
						"员工 %s 班次 '%s' 有 %d 次，偏离平均 %.1f 次",
						emp.Name, shiftType, count, deviation,
					),
					Severity: "warning",
					Penalty:  penalty,
				})
			}
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *ShiftDistributionConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// 辅助函数

// calculateStats 计算平均值和标准差
func calculateStats(values []float64) (avg, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}

	// 计算平均值
	var sum float64
	for _, v := range values {
		sum += v
	}
	avg = sum / float64(len(values))

	// 计算标准差
	var sumSquares float64
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}
	stdDev = math.Sqrt(sumSquares / float64(len(values)))

	return avg, stdDev
}

// isWeekend 判断日期是否为周末
func isWeekend(date string) bool {
	// 解析日期并检查是否为周末
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return false
	}
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

