// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// ===== 护理计划合规约束 =====

// CarePlanComplianceConstraint 护理计划合规约束
type CarePlanComplianceConstraint struct {
	*BaseConstraint
}

// NewCarePlanComplianceConstraint 创建护理计划合规约束
func NewCarePlanComplianceConstraint() *CarePlanComplianceConstraint {
	return &CarePlanComplianceConstraint{
		BaseConstraint: NewBaseConstraint(
			"护理计划合规",
			constraint.TypeCarePlanCompliance,
			constraint.CategoryHard,
			100,
		),
	}
}

// Evaluate 评估整个排班
func (c *CarePlanComplianceConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	isValid := true
	totalPenalty := 0

	for _, assignment := range ctx.Assignments {
		employee := ctx.GetEmployee(assignment.EmployeeID)
		if employee == nil {
			continue
		}

		if !c.hasRequiredNursingLevel(employee) {
			isValid = false
			penalty := c.Weight()
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     employee.ID,
				Date:           assignment.Date,
				Message:        fmt.Sprintf("护理员 %s 资质不满足护理计划要求", employee.Name),
				Severity:       "error",
				Penalty:        penalty,
			})
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *CarePlanComplianceConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	employee := ctx.GetEmployee(a.EmployeeID)
	if employee == nil {
		return true, 0
	}

	if !c.hasRequiredNursingLevel(employee) {
		return false, c.Weight()
	}

	return true, 0
}

func (c *CarePlanComplianceConstraint) hasRequiredNursingLevel(employee *model.Employee) bool {
	nursingSkills := []string{"初级护理", "中级护理", "高级护理", "护理员", "养老护理"}
	for _, skill := range employee.Skills {
		for _, nursingSkill := range nursingSkills {
			if skill == nursingSkill {
				return true
			}
		}
	}
	return len(employee.Skills) > 0
}

// ===== 护理员连续性约束 =====

// CaregiverContinuityConstraint 护理员连续性约束
type CaregiverContinuityConstraint struct {
	*BaseConstraint
}

// NewCaregiverContinuityConstraint 创建护理员连续性约束
func NewCaregiverContinuityConstraint(weight int) *CaregiverContinuityConstraint {
	return &CaregiverContinuityConstraint{
		BaseConstraint: NewBaseConstraint(
			"护理员连续性",
			constraint.TypeCaregiverContinuity,
			constraint.CategorySoft,
			weight,
		),
	}
}

// Evaluate 评估整个排班
func (c *CaregiverContinuityConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	if len(ctx.Assignments) > 1 {
		uniqueEmployees := make(map[string]bool)
		for _, a := range ctx.Assignments {
			uniqueEmployees[a.EmployeeID.String()] = true
		}

		if len(uniqueEmployees) > 1 {
			penalty := (len(uniqueEmployees) - 1) * 10
			totalPenalty = penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				Message:        fmt.Sprintf("涉及 %d 名护理员，建议减少更换频率提高连续性", len(uniqueEmployees)),
				Severity:       "warning",
				Penalty:        penalty,
			})
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *CaregiverContinuityConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// ===== 服务时间规律性约束 =====

// ServiceTimeRegularityConstraint 服务时间规律性约束
type ServiceTimeRegularityConstraint struct {
	*BaseConstraint
}

// NewServiceTimeRegularityConstraint 创建服务时间规律性约束
func NewServiceTimeRegularityConstraint(weight int) *ServiceTimeRegularityConstraint {
	return &ServiceTimeRegularityConstraint{
		BaseConstraint: NewBaseConstraint(
			"服务时间规律性",
			constraint.TypeServiceContinuity,
			constraint.CategorySoft,
			weight,
		),
	}
}

// Evaluate 评估整个排班
func (c *ServiceTimeRegularityConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	shiftSet := make(map[string]bool)
	for _, a := range ctx.Assignments {
		shiftSet[a.ShiftID.String()] = true
	}

	if len(shiftSet) > 2 {
		penalty := (len(shiftSet) - 2) * 10
		totalPenalty = penalty

		violations = append(violations, constraint.ViolationDetail{
			ConstraintType: c.Type(),
			ConstraintName: c.Name(),
			Message:        fmt.Sprintf("使用了 %d 种不同时段，建议统一服务时间提高规律性", len(shiftSet)),
			Severity:       "warning",
			Penalty:        penalty,
		})
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *ServiceTimeRegularityConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// ===== 每日最大服务患者数约束 =====

// MaxPatientsPerDayConstraint 每日最大服务患者数约束
type MaxPatientsPerDayConstraint struct {
	*BaseConstraint
	maxPatients int
}

// NewMaxPatientsPerDayConstraint 创建每日最大服务患者数约束
func NewMaxPatientsPerDayConstraint(maxPatients int) *MaxPatientsPerDayConstraint {
	return &MaxPatientsPerDayConstraint{
		BaseConstraint: NewBaseConstraint(
			"每日最大服务患者数",
			constraint.TypeMaxOrdersPerDay,
			constraint.CategoryHard,
			100,
		),
		maxPatients: maxPatients,
	}
}

// Evaluate 评估整个排班
func (c *MaxPatientsPerDayConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	isValid := true
	totalPenalty := 0

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		byDate := make(map[string]int)
		for _, a := range assignments {
			byDate[a.Date]++
		}

		for date, count := range byDate {
			if count > c.maxPatients {
				isValid = false
				penalty := (count - c.maxPatients) * c.Weight()
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           date,
					Message:        fmt.Sprintf("员工 %s 在 %s 服务 %d 位患者，超过限制 %d", emp.Name, date, count, c.maxPatients),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MaxPatientsPerDayConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	count := 0
	for _, existing := range ctx.Assignments {
		if existing.EmployeeID == a.EmployeeID && existing.Date == a.Date {
			count++
		}
	}

	if count >= c.maxPatients {
		return false, c.Weight()
	}

	return true, 0
}
