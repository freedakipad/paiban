// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// ===== 服务区域匹配约束 =====

// ServiceAreaMatchConstraint 服务区域匹配约束
type ServiceAreaMatchConstraint struct {
	*BaseConstraint
}

// NewServiceAreaMatchConstraint 创建服务区域匹配约束
func NewServiceAreaMatchConstraint() *ServiceAreaMatchConstraint {
	return &ServiceAreaMatchConstraint{
		BaseConstraint: NewBaseConstraint(
			"服务区域匹配",
			constraint.TypeServiceAreaMatch,
			constraint.CategoryHard,
			100,
		),
	}
}

// Evaluate 评估整个排班
func (c *ServiceAreaMatchConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	// 服务区域匹配：检查员工的 Extra 配置是否包含区域信息
	// 默认通过，实际使用时需根据业务逻辑扩展
	return true, 0, nil
}

// EvaluateAssignment 评估单个分配
func (c *ServiceAreaMatchConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 服务区域匹配：检查员工是否覆盖该区域
	// 默认通过
	return true, 0
}

// ===== 路程时间缓冲约束 =====

// TravelTimeBufferConstraint 路程时间缓冲约束
type TravelTimeBufferConstraint struct {
	*BaseConstraint
	minBufferMinutes int
}

// NewTravelTimeBufferConstraint 创建路程时间缓冲约束
func NewTravelTimeBufferConstraint(minBufferMinutes int) *TravelTimeBufferConstraint {
	return &TravelTimeBufferConstraint{
		BaseConstraint: NewBaseConstraint(
			"路程时间缓冲",
			constraint.TypeTravelTimeBuffer,
			constraint.CategorySoft,
			60,
		),
		minBufferMinutes: minBufferMinutes,
	}
}

// Evaluate 评估整个排班
func (c *TravelTimeBufferConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	// 按员工分组
	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		// 按日期分组
		byDate := make(map[string]int)
		for _, a := range assignments {
			byDate[a.Date]++
		}

		for date, count := range byDate {
			if count > 1 {
				penalty := (count - 1) * 5
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           date,
					Message:        fmt.Sprintf("员工 %s 在 %s 有 %d 个服务，建议增加通勤缓冲", emp.Name, date, count),
					Severity:       "warning",
					Penalty:        penalty,
				})
			}
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *TravelTimeBufferConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// ===== 客户偏好约束 =====

// CustomerPreferenceConstraint 客户偏好约束
type CustomerPreferenceConstraint struct {
	*BaseConstraint
}

// NewCustomerPreferenceConstraint 创建客户偏好约束
func NewCustomerPreferenceConstraint(weight int) *CustomerPreferenceConstraint {
	return &CustomerPreferenceConstraint{
		BaseConstraint: NewBaseConstraint(
			"客户偏好",
			constraint.TypeCustomerPreference,
			constraint.CategorySoft,
			weight,
		),
	}
}

// Evaluate 评估整个排班
func (c *CustomerPreferenceConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	return true, 0, nil
}

// EvaluateAssignment 评估单个分配
func (c *CustomerPreferenceConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}
