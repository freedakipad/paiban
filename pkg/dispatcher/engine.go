// Package dispatcher 提供智能派单引擎
package dispatcher

import (
	"log"
	"sort"

	"github.com/paiban/paiban/pkg/dispatcher/constraint"
	"github.com/paiban/paiban/pkg/model"
)

// DispatchEngine 派单引擎
type DispatchEngine struct {
	constraints []constraint.DispatchConstraint
}

// NewDispatchEngine 创建派单引擎
func NewDispatchEngine() *DispatchEngine {
	return &DispatchEngine{
		constraints: constraint.DefaultDispatchConstraints(),
	}
}

// NewDispatchEngineWithConstraints 创建带自定义约束的派单引擎
func NewDispatchEngineWithConstraints(constraints []constraint.DispatchConstraint) *DispatchEngine {
	return &DispatchEngine{
		constraints: constraints,
	}
}

// DispatchRequest 派单请求
type DispatchRequest struct {
	Order          *model.ServiceOrder
	Candidates     []*model.Employee
	Customer       *model.Customer
	TodayOrders    []*model.ServiceOrder
	ServiceHistory []model.CustomerEmployeeHistory
	MaxResults     int
}

// DispatchResponse 派单响应
type DispatchResponse struct {
	OrderID      string           `json:"order_id"`
	Success      bool             `json:"success"`
	BestMatch    *CandidateScore  `json:"best_match,omitempty"`
	Alternatives []CandidateScore `json:"alternatives,omitempty"`
	Reason       string           `json:"reason,omitempty"`
}

// CandidateScore 候选人评分
type CandidateScore struct {
	Employee     *model.Employee `json:"employee"`
	Score        float64         `json:"score"`
	Feasible     bool            `json:"feasible"`
	Violations   []string        `json:"violations,omitempty"`
	MatchReasons []string        `json:"match_reasons,omitempty"`
	Distance     float64         `json:"distance_km,omitempty"`
	TravelTime   int             `json:"travel_time_min,omitempty"`
}

// Dispatch 执行派单
func (e *DispatchEngine) Dispatch(req *DispatchRequest) *DispatchResponse {
	if req.Order == nil || len(req.Candidates) == 0 {
		return &DispatchResponse{
			Success: false,
			Reason:  "缺少订单或候选人",
		}
	}

	log.Printf("开始派单: 订单=%s, 候选人=%d", req.Order.OrderNo, len(req.Candidates))

	// 评估所有候选人
	scores := e.evaluateCandidates(req)

	// 按分数排序（分数越低越好）
	sort.Slice(scores, func(i, j int) bool {
		// 可行解优先
		if scores[i].Feasible != scores[j].Feasible {
			return scores[i].Feasible
		}
		return scores[i].Score < scores[j].Score
	})

	// 过滤可行解
	var feasibleScores []CandidateScore
	for _, s := range scores {
		if s.Feasible {
			feasibleScores = append(feasibleScores, s)
		}
	}

	// 返回结果
	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}

	if len(feasibleScores) == 0 {
		// 没有可行解
		return &DispatchResponse{
			OrderID:      req.Order.OrderNo,
			Success:      false,
			Reason:       "没有符合条件的员工",
			Alternatives: limitCandidates(scores, maxResults),
		}
	}

	// 有可行解
	response := &DispatchResponse{
		OrderID:   req.Order.OrderNo,
		Success:   true,
		BestMatch: &feasibleScores[0],
	}

	if len(feasibleScores) > 1 {
		response.Alternatives = limitCandidates(feasibleScores[1:], maxResults-1)
	}

	log.Printf("派单完成: 最佳匹配=%s, 分数=%.2f, 备选=%d",
		feasibleScores[0].Employee.Name, feasibleScores[0].Score, len(response.Alternatives))

	return response
}

// evaluateCandidates 评估所有候选人
func (e *DispatchEngine) evaluateCandidates(req *DispatchRequest) []CandidateScore {
	scores := make([]CandidateScore, 0, len(req.Candidates))

	for _, emp := range req.Candidates {
		score := e.evaluateCandidate(emp, req)
		scores = append(scores, score)
	}

	return scores
}

// evaluateCandidate 评估单个候选人
func (e *DispatchEngine) evaluateCandidate(employee *model.Employee, req *DispatchRequest) CandidateScore {
	score := CandidateScore{
		Employee: employee,
		Feasible: true,
		Score:    0,
	}

	// 获取员工今日已分配订单
	var employeeOrders []*model.ServiceOrder
	for _, order := range req.TodayOrders {
		if order.EmployeeID != nil && *order.EmployeeID == employee.ID {
			employeeOrders = append(employeeOrders, order)
		}
	}

	// 构建上下文
	ctx := &constraint.DispatchContext{
		Customer:         req.Customer,
		TodayOrders:      req.TodayOrders,
		EmployeeOrders:   employeeOrders,
		ServiceHistory:   req.ServiceHistory,
		EmployeeLocation: nil, // TODO: 获取员工位置
	}

	// 评估所有约束
	for _, c := range e.constraints {
		valid, penalty, violation := c.Evaluate(req.Order, employee, ctx)

		if !valid {
			score.Feasible = false
			score.Violations = append(score.Violations, violation)
			score.Score += penalty
		} else if penalty != 0 {
			score.Score += penalty
			if penalty < 0 {
				// 奖励转为匹配原因
				score.MatchReasons = append(score.MatchReasons, c.Name()+": 匹配")
			}
		}
	}

	return score
}

// BatchDispatch 批量派单
func (e *DispatchEngine) BatchDispatch(orders []*model.ServiceOrder, candidates []*model.Employee, customer *model.Customer) []*DispatchResponse {
	responses := make([]*DispatchResponse, len(orders))

	// 已分配的订单（用于避免时间冲突）
	assignedOrders := make([]*model.ServiceOrder, 0)

	for i, order := range orders {
		req := &DispatchRequest{
			Order:       order,
			Candidates:  candidates,
			Customer:    customer,
			TodayOrders: assignedOrders,
			MaxResults:  3,
		}

		resp := e.Dispatch(req)
		responses[i] = resp

		// 如果派单成功，记录分配
		if resp.Success && resp.BestMatch != nil {
			orderCopy := *order
			orderCopy.EmployeeID = &resp.BestMatch.Employee.ID
			orderCopy.Status = "assigned"
			assignedOrders = append(assignedOrders, &orderCopy)
		}
	}

	return responses
}

// limitCandidates 限制候选人数量
func limitCandidates(scores []CandidateScore, max int) []CandidateScore {
	if len(scores) <= max {
		return scores
	}
	return scores[:max]
}

// OptimalRoute 计算最优路线
func (e *DispatchEngine) OptimalRoute(orders []*model.ServiceOrder, startLocation *model.Location) []*model.ServiceOrder {
	if len(orders) <= 1 || startLocation == nil {
		return orders
	}

	// 简单贪心算法：每次选择最近的订单
	result := make([]*model.ServiceOrder, 0, len(orders))
	remaining := make([]*model.ServiceOrder, len(orders))
	copy(remaining, orders)

	currentLoc := *startLocation

	for len(remaining) > 0 {
		// 找最近的订单
		minDist := -1.0
		minIdx := 0

		for i, order := range remaining {
			if order.Location == nil {
				continue
			}
			dist := currentLoc.Distance(*order.Location)
			if minDist < 0 || dist < minDist {
				minDist = dist
				minIdx = i
			}
		}

		// 添加到结果
		result = append(result, remaining[minIdx])

		// 更新当前位置
		if remaining[minIdx].Location != nil {
			currentLoc = *remaining[minIdx].Location
		}

		// 移除已处理订单
		remaining = append(remaining[:minIdx], remaining[minIdx+1:]...)
	}

	return result
}
