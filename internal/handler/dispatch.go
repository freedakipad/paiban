// Package handler 提供API处理器
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/paiban/paiban/pkg/dispatcher"
	"github.com/paiban/paiban/pkg/model"
)

// DispatchRequest 派单API请求
type DispatchRequest struct {
	Order       *model.ServiceOrder              `json:"order"`
	Candidates  []*model.Employee                `json:"candidates"`
	Customer    *model.Customer                  `json:"customer,omitempty"`
	TodayOrders []*model.ServiceOrder            `json:"today_orders,omitempty"`
	History     []model.CustomerEmployeeHistory  `json:"history,omitempty"`
	MaxResults  int                              `json:"max_results,omitempty"`
}

// BatchDispatchRequest 批量派单请求
type BatchDispatchRequest struct {
	Orders     []*model.ServiceOrder `json:"orders"`
	Candidates []*model.Employee     `json:"candidates"`
	Customer   *model.Customer       `json:"customer,omitempty"`
}

// DispatchAPIResponse 派单API响应
type DispatchAPIResponse struct {
	Success     bool                       `json:"success"`
	Data        *dispatcher.DispatchResponse `json:"data,omitempty"`
	Error       string                     `json:"error,omitempty"`
}

// BatchDispatchAPIResponse 批量派单API响应
type BatchDispatchAPIResponse struct {
	Success bool                         `json:"success"`
	Data    []*dispatcher.DispatchResponse `json:"data,omitempty"`
	Summary *BatchSummary                 `json:"summary,omitempty"`
	Error   string                        `json:"error,omitempty"`
}

// BatchSummary 批量派单汇总
type BatchSummary struct {
	TotalOrders      int `json:"total_orders"`
	SuccessCount     int `json:"success_count"`
	FailCount        int `json:"fail_count"`
	AssignedEmployees int `json:"assigned_employees"`
}

var dispatchEngine *dispatcher.DispatchEngine

func init() {
	dispatchEngine = dispatcher.NewDispatchEngine()
}

// DispatchHandler 单个订单派单
func DispatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendDispatchError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Order == nil {
		sendDispatchError(w, "Order is required", http.StatusBadRequest)
		return
	}

	if len(req.Candidates) == 0 {
		sendDispatchError(w, "At least one candidate is required", http.StatusBadRequest)
		return
	}

	log.Printf("接收派单请求: order=%s, candidates=%d", req.Order.OrderNo, len(req.Candidates))

	// 构建派单请求
	dispReq := &dispatcher.DispatchRequest{
		Order:          req.Order,
		Candidates:     req.Candidates,
		Customer:       req.Customer,
		TodayOrders:    req.TodayOrders,
		ServiceHistory: req.History,
		MaxResults:     req.MaxResults,
	}

	// 执行派单
	resp := dispatchEngine.Dispatch(dispReq)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(DispatchAPIResponse{
		Success: resp.Success,
		Data:    resp,
	})
}

// BatchDispatchHandler 批量派单
func BatchDispatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BatchDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendDispatchError(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(req.Orders) == 0 {
		sendDispatchError(w, "At least one order is required", http.StatusBadRequest)
		return
	}

	if len(req.Candidates) == 0 {
		sendDispatchError(w, "At least one candidate is required", http.StatusBadRequest)
		return
	}

	log.Printf("接收批量派单请求: orders=%d, candidates=%d", len(req.Orders), len(req.Candidates))

	// 执行批量派单
	responses := dispatchEngine.BatchDispatch(req.Orders, req.Candidates, req.Customer)

	// 统计结果
	summary := &BatchSummary{
		TotalOrders: len(req.Orders),
	}
	assignedMap := make(map[string]bool)

	for _, resp := range responses {
		if resp.Success {
			summary.SuccessCount++
			if resp.BestMatch != nil {
				assignedMap[resp.BestMatch.Employee.ID.String()] = true
			}
		} else {
			summary.FailCount++
		}
	}
	summary.AssignedEmployees = len(assignedMap)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BatchDispatchAPIResponse{
		Success: true,
		Data:    responses,
		Summary: summary,
	})
}

// OptimalRouteRequest 最优路线请求
type OptimalRouteRequest struct {
	Orders        []*model.ServiceOrder `json:"orders"`
	StartLocation *model.Location       `json:"start_location"`
}

// OptimalRouteResponse 最优路线响应
type OptimalRouteResponse struct {
	Success      bool                  `json:"success"`
	Orders       []*model.ServiceOrder `json:"orders,omitempty"`
	TotalDistance float64              `json:"total_distance_km,omitempty"`
	Error        string                `json:"error,omitempty"`
}

// OptimalRouteHandler 计算最优路线
func OptimalRouteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req OptimalRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(OptimalRouteResponse{
			Success: false,
			Error:   "Invalid request: " + err.Error(),
		})
		return
	}

	if len(req.Orders) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(OptimalRouteResponse{
			Success: true,
			Orders:  req.Orders,
		})
		return
	}

	// 计算最优路线
	optimizedOrders := dispatchEngine.OptimalRoute(req.Orders, req.StartLocation)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(OptimalRouteResponse{
		Success: true,
		Orders:  optimizedOrders,
	})
}

// sendDispatchError 发送派单错误
func sendDispatchError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(DispatchAPIResponse{
		Success: false,
		Error:   message,
	})
}

