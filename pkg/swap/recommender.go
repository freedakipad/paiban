// Package swap 提供换班/调班功能
package swap

import (
	"sort"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// Recommender 换班推荐器
type Recommender struct {
	evaluator *SwapEvaluator
}

// NewRecommender 创建换班推荐器
func NewRecommender(cm *constraint.Manager) *Recommender {
	return &Recommender{
		evaluator: NewSwapEvaluator(cm),
	}
}

// Recommendation 换班推荐
type Recommendation struct {
	TargetEmployee  *model.Employee   `json:"target_employee"`
	Assignment      *model.Assignment `json:"assignment,omitempty"`
	Score           float64           `json:"score"`
	Reason          string            `json:"reason"`
	SwapType        string            `json:"swap_type"` // take_over/exchange
	ImpactSummary   string            `json:"impact_summary"`
	Rank            int               `json:"rank"`
}

// RecommendOptions 推荐选项
type RecommendOptions struct {
	MaxRecommendations int      // 最大推荐数量
	PreferredEmployees []uuid.UUID // 优先考虑的员工
	ExcludeEmployees   []uuid.UUID // 排除的员工
	AllowExchange      bool     // 是否允许互换
	MinScore           float64  // 最低得分
}

// DefaultRecommendOptions 返回默认选项
func DefaultRecommendOptions() *RecommendOptions {
	return &RecommendOptions{
		MaxRecommendations: 5,
		AllowExchange:      true,
		MinScore:           60,
	}
}

// RecommendSwapTargets 推荐换班目标员工
func (r *Recommender) RecommendSwapTargets(
	ctx *constraint.Context,
	sourceAssignment *model.Assignment,
	options *RecommendOptions,
) []Recommendation {
	if options == nil {
		options = DefaultRecommendOptions()
	}

	var candidates []Recommendation

	// 排除原员工
	excludeSet := make(map[uuid.UUID]bool)
	excludeSet[sourceAssignment.EmployeeID] = true
	for _, id := range options.ExcludeEmployees {
		excludeSet[id] = true
	}

	// 优先员工集合
	preferredSet := make(map[uuid.UUID]bool)
	for _, id := range options.PreferredEmployees {
		preferredSet[id] = true
	}

	// 遍历所有员工评估
	for _, emp := range ctx.Employees {
		if excludeSet[emp.ID] {
			continue
		}
		if !emp.IsActive() {
			continue
		}

		// 评估接管换班
		evaluation := r.evaluator.EvaluateSwap(ctx, &SwapRequest{
			SourceAssignment: sourceAssignment,
			TargetEmployee:   emp,
		})

		if !evaluation.Feasible {
			continue
		}

		if evaluation.Score < options.MinScore {
			continue
		}

		candidate := Recommendation{
			TargetEmployee: emp,
			Score:          evaluation.Score,
			SwapType:       "take_over",
			Reason:         r.generateReason(emp, evaluation),
			ImpactSummary:  r.generateImpactSummary(evaluation),
		}

		// 优先员工加分
		if preferredSet[emp.ID] {
			candidate.Score += 10
		}

		candidates = append(candidates, candidate)

		// 如果允许互换，检查该员工当天的班次
		if options.AllowExchange {
			exchangeCandidates := r.findExchangeCandidates(ctx, sourceAssignment, emp, options)
			candidates = append(candidates, exchangeCandidates...)
		}
	}

	// 排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// 截取结果
	if len(candidates) > options.MaxRecommendations {
		candidates = candidates[:options.MaxRecommendations]
	}

	// 设置排名
	for i := range candidates {
		candidates[i].Rank = i + 1
	}

	return candidates
}

// findExchangeCandidates 查找互换候选
func (r *Recommender) findExchangeCandidates(
	ctx *constraint.Context,
	sourceAssignment *model.Assignment,
	targetEmp *model.Employee,
	options *RecommendOptions,
) []Recommendation {
	var candidates []Recommendation

	// 查找目标员工的排班
	targetAssignments := ctx.GetEmployeeAssignments(targetEmp.ID)

	for _, targetAss := range targetAssignments {
		// 跳过同一天的排班（避免同一天互换）
		if targetAss.Date == sourceAssignment.Date {
			continue
		}

		// 评估互换
		evaluation := r.evaluator.EvaluateSwap(ctx, &SwapRequest{
			SourceAssignment: sourceAssignment,
			TargetEmployee:   targetEmp,
			TargetAssignment: targetAss,
		})

		if !evaluation.Feasible {
			continue
		}

		if evaluation.Score < options.MinScore {
			continue
		}

		candidates = append(candidates, Recommendation{
			TargetEmployee: targetEmp,
			Assignment:     targetAss,
			Score:          evaluation.Score,
			SwapType:       "exchange",
			Reason:         "互换班次，双方工时更平衡",
			ImpactSummary:  r.generateImpactSummary(evaluation),
		})
	}

	return candidates
}

// generateReason 生成推荐原因
func (r *Recommender) generateReason(emp *model.Employee, evaluation *SwapEvaluation) string {
	reasons := []string{}

	// 检查工时情况
	if evaluation.Impact != nil && evaluation.Impact.TargetEmployeeImpact != nil {
		if evaluation.Impact.TargetEmployeeImpact.HoursChange <= 8 {
			reasons = append(reasons, "工时增加合理")
		}
	}

	// 检查问题数量
	if len(evaluation.Issues) == 0 {
		reasons = append(reasons, "无约束冲突")
	} else {
		warningCount := 0
		for _, issue := range evaluation.Issues {
			if issue.Severity == "warning" {
				warningCount++
			}
		}
		if warningCount > 0 && warningCount <= 2 {
			reasons = append(reasons, "仅有少量软约束提醒")
		}
	}

	// 技能匹配
	reasons = append(reasons, "技能匹配")

	if len(reasons) == 0 {
		return "可以接替此班次"
	}

	return reasons[0]
}

// generateImpactSummary 生成影响摘要
func (r *Recommender) generateImpactSummary(evaluation *SwapEvaluation) string {
	if evaluation.Impact == nil {
		return "影响较小"
	}

	targetImpact := evaluation.Impact.TargetEmployeeImpact
	if targetImpact == nil {
		return "影响较小"
	}

	if targetImpact.HoursChange > 0 {
		return "目标员工增加工时，更接近平均水平"
	} else if targetImpact.HoursChange < 0 {
		return "目标员工减少工时"
	}

	return "对双方工时影响均衡"
}

// FindBestSwapMatch 为请假员工找到最佳替换
func (r *Recommender) FindBestSwapMatch(
	ctx *constraint.Context,
	employeeID uuid.UUID,
	date string,
) *Recommendation {
	// 找到该员工当天的排班
	assignments := ctx.GetEmployeeAssignments(employeeID)
	var targetAssignment *model.Assignment

	for _, a := range assignments {
		if a.Date == date {
			targetAssignment = a
			break
		}
	}

	if targetAssignment == nil {
		return nil
	}

	// 获取推荐
	recommendations := r.RecommendSwapTargets(ctx, targetAssignment, &RecommendOptions{
		MaxRecommendations: 1,
		MinScore:           50,
	})

	if len(recommendations) == 0 {
		return nil
	}

	return &recommendations[0]
}

// AutoAssignSwap 自动分配换班
func (r *Recommender) AutoAssignSwap(
	ctx *constraint.Context,
	sourceAssignment *model.Assignment,
) (*model.Assignment, error) {
	recommendations := r.RecommendSwapTargets(ctx, sourceAssignment, &RecommendOptions{
		MaxRecommendations: 1,
		MinScore:           70, // 自动分配要求更高得分
	})

	if len(recommendations) == 0 {
		return nil, nil
	}

	best := recommendations[0]

	// 创建新的分配
	newAssignment := *sourceAssignment
	newAssignment.ID = uuid.New()
	newAssignment.EmployeeID = best.TargetEmployee.ID
	newAssignment.IsSwapped = true
	newAssignment.OriginalEmpID = &sourceAssignment.EmployeeID

	return &newAssignment, nil
}

