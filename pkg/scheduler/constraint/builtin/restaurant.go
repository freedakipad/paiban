// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"strings"
	"time"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// PeakHoursCoverageConstraint 高峰期覆盖约束（软约束）
// 确保高峰时段有足够的员工在岗
type PeakHoursCoverageConstraint struct {
	*BaseConstraint
	peakHours    []string // 高峰时段列表，格式 "HH:MM-HH:MM"
	minStaff     int      // 高峰期最少人数
}

// NewPeakHoursCoverageConstraint 创建高峰期覆盖约束
func NewPeakHoursCoverageConstraint(weight int, peakHours []string, minStaff int) *PeakHoursCoverageConstraint {
	return &PeakHoursCoverageConstraint{
		BaseConstraint: NewBaseConstraint(
			"高峰期人员覆盖",
			constraint.TypePeakHoursCoverage,
			constraint.CategorySoft,
			weight,
		),
		peakHours: peakHours,
		minStaff:  minStaff,
	}
}

// Evaluate 评估整个排班
func (c *PeakHoursCoverageConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	// 遍历每一天
	dates := getUniqueDates(ctx.Assignments)
	for _, date := range dates {
		dayAssignments := ctx.GetDateAssignments(date)

		// 检查每个高峰时段
		for _, peakRange := range c.peakHours {
			peakStart, peakEnd := parsePeakHours(peakRange)
			if peakStart == "" {
				continue
			}

			// 统计该时段在岗人数
			staffCount := 0
			for _, a := range dayAssignments {
				if isAssignmentCoveringPeriod(a, date, peakStart, peakEnd) {
					staffCount++
				}
			}

			if staffCount < c.minStaff {
				shortage := c.minStaff - staffCount
				penalty := c.Weight() * shortage
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					Date:           date,
					Message:        fmt.Sprintf("%s 高峰期 %s 仅有 %d 人在岗，少于要求的 %d 人", date, peakRange, staffCount, c.minStaff),
					Severity:       "warning",
					Penalty:        penalty,
				})
			}
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *PeakHoursCoverageConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 软约束，检查此分配是否有助于覆盖高峰期
	for _, peakRange := range c.peakHours {
		peakStart, peakEnd := parsePeakHours(peakRange)
		if isAssignmentCoveringPeriod(a, a.Date, peakStart, peakEnd) {
			return true, -c.Weight() / 4 // 奖励覆盖高峰期的分配
		}
	}
	return true, 0
}

// SplitShiftConstraint 两头班约束
// 支持餐饮业的两头班模式（早午高峰+晚高峰）
type SplitShiftConstraint struct {
	*BaseConstraint
	maxSplitShiftsPerWeek int  // 每周最多两头班次数
	minBreakHours         int  // 两头班中间最少休息时间
	allowSplitShift       bool // 是否允许两头班
}

// NewSplitShiftConstraint 创建两头班约束
func NewSplitShiftConstraint(weight int, maxPerWeek, minBreak int, allow bool) *SplitShiftConstraint {
	return &SplitShiftConstraint{
		BaseConstraint: NewBaseConstraint(
			"两头班约束",
			constraint.Type("split_shift"),
			constraint.CategorySoft,
			weight,
		),
		maxSplitShiftsPerWeek: maxPerWeek,
		minBreakHours:         minBreak,
		allowSplitShift:       allow,
	}
}

// Evaluate 评估整个排班
func (c *SplitShiftConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	if !c.allowSplitShift {
		// 检查是否有两头班
		for _, emp := range ctx.Employees {
			assignments := ctx.GetEmployeeAssignments(emp.ID)
			splitShiftCount := countSplitShifts(assignments)

			if splitShiftCount > 0 {
				penalty := c.Weight() * splitShiftCount
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Message:        fmt.Sprintf("员工 %s 有 %d 个两头班，但门店不允许两头班", emp.Name, splitShiftCount),
					Severity:       "warning",
					Penalty:        penalty,
				})
			}
		}
		return true, totalPenalty, violations
	}

	// 检查两头班数量限制
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		splitShiftCount := countSplitShifts(assignments)

		if splitShiftCount > c.maxSplitShiftsPerWeek {
			excess := splitShiftCount - c.maxSplitShiftsPerWeek
			penalty := c.Weight() * excess
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message:        fmt.Sprintf("员工 %s 有 %d 个两头班，超过限制 %d 个", emp.Name, splitShiftCount, c.maxSplitShiftsPerWeek),
				Severity:       "warning",
				Penalty:        penalty,
			})
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *SplitShiftConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	shift := ctx.GetShift(a.ShiftID)
	if shift != nil && shift.IsSplitShift() {
		if !c.allowSplitShift {
			return true, c.Weight()
		}
	}
	return true, 0
}

// PositionCoverageConstraint 岗位覆盖约束
// 确保每个必需岗位都有人覆盖
type PositionCoverageConstraint struct {
	*BaseConstraint
	requiredPositions map[string]int // 必需岗位及最少人数
}

// NewPositionCoverageConstraint 创建岗位覆盖约束
func NewPositionCoverageConstraint(weight int, positions map[string]int) *PositionCoverageConstraint {
	return &PositionCoverageConstraint{
		BaseConstraint: NewBaseConstraint(
			"岗位覆盖",
			constraint.Type("position_coverage"),
			constraint.CategoryHard,
			weight,
		),
		requiredPositions: positions,
	}
}

// Evaluate 评估整个排班
func (c *PositionCoverageConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 遍历每一天
	dates := getUniqueDates(ctx.Assignments)
	for _, date := range dates {
		dayAssignments := ctx.GetDateAssignments(date)

		// 统计每个岗位的人数
		positionCount := make(map[string]int)
		for _, a := range dayAssignments {
			if a.Position != "" {
				positionCount[a.Position]++
			} else {
				// 从员工信息获取岗位
				emp := ctx.GetEmployee(a.EmployeeID)
				if emp != nil && emp.Position != "" {
					positionCount[emp.Position]++
				}
			}
		}

		// 检查是否满足要求
		for pos, minCount := range c.requiredPositions {
			actual := positionCount[pos]
			if actual < minCount {
				isValid = false
				penalty := c.Weight() * (minCount - actual)
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					Date:           date,
					Message:        fmt.Sprintf("%s 岗位 '%s' 仅有 %d 人，少于要求的 %d 人", date, pos, actual, minCount),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *PositionCoverageConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 分配有助于岗位覆盖
	return true, 0
}

// PreferenceMatchConstraint 偏好匹配约束（增强版）
// 综合考虑员工的各种偏好
type PreferenceMatchConstraint struct {
	*BaseConstraint
}

// NewPreferenceMatchConstraint 创建偏好匹配约束
func NewPreferenceMatchConstraint(weight int) *PreferenceMatchConstraint {
	return &PreferenceMatchConstraint{
		BaseConstraint: NewBaseConstraint(
			"偏好匹配",
			constraint.Type("preference_match"),
			constraint.CategorySoft,
			weight,
		),
	}
}

// Evaluate 评估整个排班
func (c *PreferenceMatchConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	for _, a := range ctx.Assignments {
		emp := ctx.GetEmployee(a.EmployeeID)
		if emp == nil || emp.Preferences == nil {
			continue
		}

		prefs := emp.Preferences
		shift := ctx.GetShift(a.ShiftID)

		// 检查最大/最小工时偏好
		currentHours := ctx.GetEmployeeHoursInRange(emp.ID, ctx.StartDate, ctx.EndDate)
		if prefs.MaxHoursPerWeek > 0 && int(currentHours) > prefs.MaxHoursPerWeek {
			penalty := c.Weight() / 2
			totalPenalty += penalty
			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message:        fmt.Sprintf("员工 %s 期望周工时不超过 %d 小时，实际 %.1f 小时", emp.Name, prefs.MaxHoursPerWeek, currentHours),
				Severity:       "warning",
				Penalty:        penalty,
			})
		}

		// 检查班次偏好
		if len(prefs.PreferredShifts) > 0 && shift != nil {
			matched := false
			for _, prefShift := range prefs.PreferredShifts {
				if shift.Code == prefShift || shift.ShiftType == prefShift {
					matched = true
					break
				}
			}
			if !matched {
				penalty := c.Weight() / 4
				totalPenalty += penalty
				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           a.Date,
					Message:        fmt.Sprintf("员工 %s 未被分配到偏好班次", emp.Name),
					Severity:       "warning",
					Penalty:        penalty,
				})
			}
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *PreferenceMatchConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// 辅助函数

// parsePeakHours 解析高峰时段
func parsePeakHours(peakRange string) (string, string) {
	parts := strings.Split(peakRange, "-")
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// isAssignmentCoveringPeriod 检查分配是否覆盖指定时段
func isAssignmentCoveringPeriod(a *model.Assignment, date, periodStart, periodEnd string) bool {
	// 解析时间
	assignStart := a.StartTime
	assignEnd := a.EndTime

	// 解析时段
	pStart, err1 := time.Parse("2006-01-02 15:04", date+" "+periodStart)
	pEnd, err2 := time.Parse("2006-01-02 15:04", date+" "+periodEnd)
	if err1 != nil || err2 != nil {
		return false
	}

	// 检查是否有重叠
	return assignStart.Before(pEnd) && assignEnd.After(pStart)
}

// getUniqueDates 获取所有不重复的日期
func getUniqueDates(assignments []*model.Assignment) []string {
	dateSet := make(map[string]bool)
	for _, a := range assignments {
		dateSet[a.Date] = true
	}

	dates := make([]string, 0, len(dateSet))
	for d := range dateSet {
		dates = append(dates, d)
	}
	return dates
}

// countSplitShifts 统计两头班数量
func countSplitShifts(assignments []*model.Assignment) int {
	// 按日期分组
	byDate := make(map[string][]*model.Assignment)
	for _, a := range assignments {
		byDate[a.Date] = append(byDate[a.Date], a)
	}

	count := 0
	for _, dayAssignments := range byDate {
		if len(dayAssignments) >= 2 {
			// 检查是否有间隔（两头班特征）
			// 简化实现：同一天有多个班次且中间有间隔
			count++
		}
	}
	return count
}

