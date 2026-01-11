// Package swap 提供换班/调班功能
package swap

import (
	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
	"github.com/paiban/paiban/pkg/validator"
)

// SwapEvaluator 换班评估器
type SwapEvaluator struct {
	constraintManager *constraint.Manager
	conflictDetector  *validator.ConflictDetector
}

// NewSwapEvaluator 创建换班评估器
func NewSwapEvaluator(cm *constraint.Manager) *SwapEvaluator {
	return &SwapEvaluator{
		constraintManager: cm,
		conflictDetector:  validator.NewConflictDetector(nil),
	}
}

// SwapRequest 换班请求
type SwapRequest struct {
	SourceAssignment *model.Assignment `json:"source_assignment"`
	TargetEmployee   *model.Employee   `json:"target_employee"`
	TargetAssignment *model.Assignment `json:"target_assignment,omitempty"` // 互换时的目标班次
}

// SwapEvaluation 换班评估结果
type SwapEvaluation struct {
	Feasible       bool        `json:"feasible"`
	Score          float64     `json:"score"`  // 0-100
	Issues         []SwapIssue `json:"issues"` // 问题列表
	Impact         *SwapImpact `json:"impact"` // 影响分析
	Recommendation string      `json:"recommendation"`
}

// SwapIssue 换班问题
type SwapIssue struct {
	Type     string `json:"type"`
	Severity string `json:"severity"` // error/warning/info
	Message  string `json:"message"`
}

// SwapImpact 换班影响
type SwapImpact struct {
	SourceEmployeeImpact *EmployeeImpact `json:"source_employee_impact"`
	TargetEmployeeImpact *EmployeeImpact `json:"target_employee_impact"`
	OverallScoreChange   float64         `json:"overall_score_change"`
}

// EmployeeImpact 员工影响
type EmployeeImpact struct {
	HoursChange         float64 `json:"hours_change"`
	OvertimeChange      float64 `json:"overtime_change"`
	PreferenceSatisfied bool    `json:"preference_satisfied"`
	NewConflicts        int     `json:"new_conflicts"`
}

// EvaluateSwap 评估换班可行性
func (e *SwapEvaluator) EvaluateSwap(
	ctx *constraint.Context,
	request *SwapRequest,
) *SwapEvaluation {
	result := &SwapEvaluation{
		Feasible: true,
		Score:    100,
		Issues:   make([]SwapIssue, 0),
		Impact: &SwapImpact{
			SourceEmployeeImpact: &EmployeeImpact{},
			TargetEmployeeImpact: &EmployeeImpact{},
		},
	}

	source := request.SourceAssignment
	targetEmp := request.TargetEmployee

	// 1. 基础检查
	if source == nil || targetEmp == nil {
		result.Feasible = false
		result.Issues = append(result.Issues, SwapIssue{
			Type:     "invalid_request",
			Severity: "error",
			Message:  "无效的换班请求",
		})
		return result
	}

	// 2. 检查目标员工是否在职
	if !targetEmp.IsActive() {
		result.Feasible = false
		result.Issues = append(result.Issues, SwapIssue{
			Type:     "employee_inactive",
			Severity: "error",
			Message:  "目标员工不在职",
		})
		return result
	}

	// 3. 检查技能匹配
	shift := ctx.GetShift(source.ShiftID)
	if shift != nil {
		// 查找对应需求
		for _, req := range ctx.Requirements {
			if req.ShiftID == source.ShiftID && req.Date == source.Date {
				for _, skill := range req.Skills {
					if !targetEmp.HasSkill(skill) {
						result.Feasible = false
						result.Issues = append(result.Issues, SwapIssue{
							Type:     "skill_mismatch",
							Severity: "error",
							Message:  "目标员工缺少必需技能: " + skill,
						})
					}
				}
			}
		}
	}

	// 4. 模拟换班后检测冲突
	simulatedAssignments := e.simulateSwap(ctx, request)
	employees := make(map[uuid.UUID]*model.Employee)
	for _, emp := range ctx.Employees {
		employees[emp.ID] = emp
	}

	conflicts := e.conflictDetector.DetectAll(simulatedAssignments, employees)
	for _, conflict := range conflicts {
		if conflict.EmployeeID == targetEmp.ID {
			result.Issues = append(result.Issues, SwapIssue{
				Type:     string(conflict.Type),
				Severity: conflict.Severity,
				Message:  conflict.Message,
			})
			if conflict.Severity == "error" {
				result.Feasible = false
			}
		}
	}

	// 5. 使用约束管理器评估
	if e.constraintManager != nil {
		// 创建模拟上下文
		simCtx := e.createSimulatedContext(ctx, request)
		constraintResult := e.constraintManager.Evaluate(simCtx)

		if !constraintResult.IsValid {
			for _, v := range constraintResult.HardViolations {
				if v.EmployeeID == targetEmp.ID {
					result.Feasible = false
					result.Issues = append(result.Issues, SwapIssue{
						Type:     string(v.ConstraintType),
						Severity: "error",
						Message:  v.Message,
					})
				}
			}
		}

		// 更新得分
		result.Score = constraintResult.Score
	}

	// 6. 计算影响
	e.calculateImpact(ctx, request, result)

	// 7. 生成建议
	result.Recommendation = e.generateRecommendation(result)

	return result
}

// simulateSwap 模拟换班后的排班
func (e *SwapEvaluator) simulateSwap(ctx *constraint.Context, request *SwapRequest) []*model.Assignment {
	var simulated []*model.Assignment

	for _, a := range ctx.Assignments {
		if a.ID == request.SourceAssignment.ID {
			// 替换为目标员工
			newAssignment := *a
			newAssignment.EmployeeID = request.TargetEmployee.ID
			simulated = append(simulated, &newAssignment)
		} else if request.TargetAssignment != nil && a.ID == request.TargetAssignment.ID {
			// 互换场景：目标班次分配给源员工
			newAssignment := *a
			newAssignment.EmployeeID = request.SourceAssignment.EmployeeID
			simulated = append(simulated, &newAssignment)
		} else {
			simulated = append(simulated, a)
		}
	}

	return simulated
}

// createSimulatedContext 创建模拟上下文
func (e *SwapEvaluator) createSimulatedContext(ctx *constraint.Context, request *SwapRequest) *constraint.Context {
	simCtx := constraint.NewContext(ctx.OrgID, ctx.StartDate, ctx.EndDate)
	simCtx.SetEmployees(ctx.Employees)
	simCtx.SetShifts(ctx.Shifts)
	simCtx.Requirements = ctx.Requirements

	simulated := e.simulateSwap(ctx, request)
	simCtx.SetAssignments(simulated)

	return simCtx
}

// calculateImpact 计算换班影响
func (e *SwapEvaluator) calculateImpact(ctx *constraint.Context, request *SwapRequest, result *SwapEvaluation) {
	source := request.SourceAssignment
	targetEmp := request.TargetEmployee
	sourceEmp := ctx.GetEmployee(source.EmployeeID)

	if sourceEmp == nil || targetEmp == nil {
		return
	}

	// 源员工影响
	sourceCurrentHours := ctx.GetEmployeeHoursInRange(sourceEmp.ID, ctx.StartDate, ctx.EndDate)
	sourceNewHours := sourceCurrentHours - source.WorkingHours()
	result.Impact.SourceEmployeeImpact.HoursChange = sourceNewHours - sourceCurrentHours

	// 目标员工影响
	targetCurrentHours := ctx.GetEmployeeHoursInRange(targetEmp.ID, ctx.StartDate, ctx.EndDate)
	targetNewHours := targetCurrentHours + source.WorkingHours()
	result.Impact.TargetEmployeeImpact.HoursChange = targetNewHours - targetCurrentHours

	// 加班变化（假设标准40小时）
	const standardHours = 40.0
	if sourceCurrentHours > standardHours && sourceNewHours <= standardHours {
		result.Impact.SourceEmployeeImpact.OvertimeChange = sourceNewHours - sourceCurrentHours
	}
	if targetCurrentHours <= standardHours && targetNewHours > standardHours {
		result.Impact.TargetEmployeeImpact.OvertimeChange = targetNewHours - standardHours
	}
}

// generateRecommendation 生成换班建议
func (e *SwapEvaluator) generateRecommendation(result *SwapEvaluation) string {
	if !result.Feasible {
		return "不建议进行此换班，存在硬约束冲突"
	}

	if result.Score >= 90 {
		return "强烈推荐，换班后整体效果良好"
	} else if result.Score >= 70 {
		return "可以进行，但存在一些软约束问题"
	} else if result.Score >= 50 {
		return "谨慎进行，可能影响整体排班质量"
	} else {
		return "不推荐，虽然可行但会显著降低排班质量"
	}
}

// CanSwap 快速检查是否可换班
func (e *SwapEvaluator) CanSwap(ctx *constraint.Context, request *SwapRequest) (bool, string) {
	result := e.EvaluateSwap(ctx, request)
	if !result.Feasible {
		if len(result.Issues) > 0 {
			return false, result.Issues[0].Message
		}
		return false, "无法进行换班"
	}
	return true, ""
}
