// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// SkillRequiredConstraint 技能要求约束
type SkillRequiredConstraint struct {
	*BaseConstraint
}

// NewSkillRequiredConstraint 创建技能要求约束
func NewSkillRequiredConstraint() *SkillRequiredConstraint {
	return &SkillRequiredConstraint{
		BaseConstraint: NewBaseConstraint(
			"技能要求",
			constraint.TypeSkillRequired,
			constraint.CategoryHard,
			100,
		),
	}
}

// Evaluate 评估整个排班
func (c *SkillRequiredConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	// 遍历所有分配，检查技能匹配
	for _, a := range ctx.Assignments {
		emp := ctx.GetEmployee(a.EmployeeID)
		if emp == nil {
			continue
		}

		// 获取班次需求的技能
		shift := ctx.GetShift(a.ShiftID)
		if shift == nil {
			continue
		}

		// 检查对应需求 - 只需要匹配至少一个需求即可
		matchedReq := false
		for _, req := range ctx.Requirements {
			if req.ShiftID != a.ShiftID || req.Date != a.Date {
				continue
			}

			// 如果分配有Position，只匹配相同Position的需求
			if a.Position != "" && req.Position != "" && a.Position != req.Position {
				continue
			}

			// 检查岗位匹配
			if req.Position != "" && emp.Position != req.Position {
				continue
			}

			// 检查技能匹配
			skillMatch := true
			for _, requiredSkill := range req.Skills {
				if !emp.HasSkill(requiredSkill) {
					skillMatch = false
					isValid = false
					penalty := c.Weight()
					totalPenalty += penalty

					violations = append(violations, constraint.ViolationDetail{
						ConstraintType: c.Type(),
						ConstraintName: c.Name(),
						EmployeeID:     emp.ID,
						Date:           a.Date,
						Message: fmt.Sprintf(
							"员工 %s 缺少必需技能: %s",
							emp.Name, requiredSkill,
						),
						Severity: "error",
						Penalty:  penalty,
					})
					break
				}
			}

			if skillMatch {
				matchedReq = true
				break // 找到匹配的需求，不再检查其他需求
			}
		}

		// 如果没有匹配的需求，检查分配是否应该存在
		if !matchedReq && a.Position != "" {
			// 检查是否有任何需求期望这个位置
			hasMatchingReq := false
			for _, req := range ctx.Requirements {
				if req.ShiftID == a.ShiftID && req.Date == a.Date && req.Position == a.Position {
					hasMatchingReq = true
					break
				}
			}
			if !hasMatchingReq {
				isValid = false
				penalty := c.Weight()
				totalPenalty += penalty
				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           a.Date,
					Message: fmt.Sprintf(
						"员工 %s 岗位 %s 没有对应需求",
						emp.Name, a.Position,
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
func (c *SkillRequiredConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	emp := ctx.GetEmployee(a.EmployeeID)
	if emp == nil {
		return false, c.Weight()
	}

	// 查找对应需求 - 只需要匹配至少一个需求即可
	matchedRequirement := false
	for _, req := range ctx.Requirements {
		if req.ShiftID != a.ShiftID || req.Date != a.Date {
			continue
		}

		// 检查岗位匹配（如果分配有position字段，需要匹配）
		if a.Position != "" && req.Position != "" && a.Position != req.Position {
			continue // 这个需求不匹配，尝试下一个
		}
		if req.Position != "" && emp.Position != req.Position {
			continue // 员工岗位不匹配这个需求，尝试下一个
		}

		// 检查所有必需技能
		skillMatch := true
		for _, skill := range req.Skills {
			if !emp.HasSkill(skill) {
				skillMatch = false
				break
			}
		}
		if !skillMatch {
			continue // 技能不匹配，尝试下一个需求
		}

		// 找到一个匹配的需求
		matchedRequirement = true
		break
	}

	if !matchedRequirement {
		// 如果没有找到匹配的需求，检查是否有任何需求对这个班次和日期有要求
		hasRequirement := false
		for _, req := range ctx.Requirements {
			if req.ShiftID == a.ShiftID && req.Date == a.Date {
				hasRequirement = true
				break
			}
		}
		if hasRequirement {
			return false, c.Weight()
		}
	}

	return true, 0
}

// WorkloadBalanceConstraint 工作量均衡约束（软约束）
type WorkloadBalanceConstraint struct {
	*BaseConstraint
	tolerancePercent float64 // 允许的偏差百分比
}

// NewWorkloadBalanceConstraint 创建工作量均衡约束
func NewWorkloadBalanceConstraint(weight int, tolerancePercent float64) *WorkloadBalanceConstraint {
	return &WorkloadBalanceConstraint{
		BaseConstraint: NewBaseConstraint(
			"工作量均衡",
			constraint.TypeWorkloadBalance,
			constraint.CategorySoft,
			weight,
		),
		tolerancePercent: tolerancePercent,
	}
}

// Evaluate 评估整个排班
func (c *WorkloadBalanceConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	if len(ctx.Employees) == 0 {
		return true, 0, nil
	}

	// 计算每个员工的工作时长
	hoursPerEmployee := make(map[string]float64)
	var totalHours float64

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		var hours float64
		for _, a := range assignments {
			hours += a.WorkingHours()
		}
		hoursPerEmployee[emp.ID.String()] = hours
		totalHours += hours
	}

	// 计算平均工时
	avgHours := totalHours / float64(len(ctx.Employees))
	tolerance := avgHours * c.tolerancePercent / 100

	// 检查每个员工的偏差
	for _, emp := range ctx.Employees {
		hours := hoursPerEmployee[emp.ID.String()]
		deviation := hours - avgHours

		if deviation > tolerance || deviation < -tolerance {
			penalty := int(abs(deviation) * float64(c.Weight()) / avgHours)
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message: fmt.Sprintf(
					"员工 %s 工时 %.1f 小时，偏离平均 %.1f 小时 (平均: %.1f)",
					emp.Name, hours, deviation, avgHours,
				),
				Severity: "warning",
				Penalty:  penalty,
			})
		}
	}

	// 软约束不影响有效性
	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *WorkloadBalanceConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 软约束，总是允许分配
	// 但返回相应的惩罚值

	if len(ctx.Employees) == 0 {
		return true, 0
	}

	// 计算当前平均工时
	var totalHours float64
	for _, emp := range ctx.Employees {
		totalHours += ctx.GetEmployeeHoursInRange(emp.ID, ctx.StartDate, ctx.EndDate)
	}
	avgHours := totalHours / float64(len(ctx.Employees))

	// 计算该员工加上新分配后的工时
	currentHours := ctx.GetEmployeeHoursInRange(a.EmployeeID, ctx.StartDate, ctx.EndDate)
	newHours := currentHours + a.WorkingHours()

	deviation := newHours - avgHours
	tolerance := avgHours * c.tolerancePercent / 100

	if deviation > tolerance {
		penalty := int(deviation * float64(c.Weight()) / (avgHours + 1))
		return true, penalty
	}

	return true, 0
}

// abs 返回绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

