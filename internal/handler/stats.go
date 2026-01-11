// Package handler 提供API处理器
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/stats"
)

// StatsRequest 统计请求
type StatsRequest struct {
	OrgID       string               `json:"org_id"`
	StartDate   string               `json:"start_date"`
	EndDate     string               `json:"end_date"`
	Employees   []*model.Employee    `json:"employees"`
	Shifts      []*model.Shift       `json:"shifts"`
	Assignments []*model.Assignment  `json:"assignments"`
}

// FairnessResponse 公平性响应
type FairnessResponse struct {
	Success bool                    `json:"success"`
	Data    *stats.FairnessMetrics  `json:"data,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// CoverageResponse 覆盖率响应
type CoverageResponse struct {
	Success bool                    `json:"success"`
	Data    *stats.CoverageMetrics  `json:"data,omitempty"`
	Error   string                  `json:"error,omitempty"`
}

// WorkloadResponse 工作量响应
type WorkloadResponse struct {
	Success bool              `json:"success"`
	Data    *WorkloadSummary  `json:"data,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// WorkloadSummary 工作量汇总
type WorkloadSummary struct {
	Period           string                 `json:"period"`
	TotalHours       float64                `json:"total_hours"`
	TotalShifts      int                    `json:"total_shifts"`
	EmployeeCount    int                    `json:"employee_count"`
	AvgHoursPerPerson float64               `json:"avg_hours_per_person"`
	OvertimeHours    float64                `json:"overtime_hours"`
	ByEmployee       []EmployeeWorkload     `json:"by_employee"`
	ByDate           map[string]DailyWorkload `json:"by_date"`
	ByShiftType      map[string]float64     `json:"by_shift_type"`
}

// EmployeeWorkload 员工工作量
type EmployeeWorkload struct {
	EmployeeID    string  `json:"employee_id"`
	EmployeeName  string  `json:"employee_name"`
	TotalHours    float64 `json:"total_hours"`
	ShiftCount    int     `json:"shift_count"`
	OvertimeHours float64 `json:"overtime_hours"`
	Utilization   float64 `json:"utilization"` // 利用率 (%)
}

// DailyWorkload 每日工作量
type DailyWorkload struct {
	Date        string  `json:"date"`
	TotalHours  float64 `json:"total_hours"`
	ShiftCount  int     `json:"shift_count"`
	StaffCount  int     `json:"staff_count"`
}

// GetFairnessHandler 公平性分析API
func GetFairnessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("接收公平性分析请求: org_id=%s, employees=%d, assignments=%d",
		req.OrgID, len(req.Employees), len(req.Assignments))

	// 转换为stats包的类型
	assignments := convertToAssignmentInfo(req.Assignments)
	employees := convertToEmployeeInfo(req.Employees)

	analyzer := stats.NewFairnessAnalyzer()
	metrics := analyzer.Analyze(assignments, employees)

	resp := FairnessResponse{
		Success: true,
		Data:    metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetCoverageHandler 覆盖率分析API
func GetCoverageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("接收覆盖率分析请求: org_id=%s, shifts=%d, assignments=%d",
		req.OrgID, len(req.Shifts), len(req.Assignments))

	// 转换为stats包的类型
	shifts := convertToShiftInfo(req.Shifts)
	assignments := convertToAssignmentInfo(req.Assignments)

	analyzer := stats.NewCoverageAnalyzer()
	metrics := analyzer.Analyze(shifts, assignments)

	resp := CoverageResponse{
		Success: true,
		Data:    metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetWorkloadHandler 工作量统计API
func GetWorkloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("接收工作量统计请求: org_id=%s, start_date=%s, end_date=%s",
		req.OrgID, req.StartDate, req.EndDate)

	// 构建员工映射
	employeeMap := make(map[string]*model.Employee)
	for _, e := range req.Employees {
		employeeMap[e.ID.String()] = e
	}

	// 计算工作量
	summary := calculateWorkload(req.Assignments, employeeMap, req.StartDate, req.EndDate)

	resp := WorkloadResponse{
		Success: true,
		Data:    summary,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// calculateWorkload 计算工作量
func calculateWorkload(assignments []*model.Assignment, employeeMap map[string]*model.Employee, startDate, endDate string) *WorkloadSummary {
	summary := &WorkloadSummary{
		Period:      startDate + " ~ " + endDate,
		ByDate:      make(map[string]DailyWorkload),
		ByShiftType: make(map[string]float64),
	}

	// 员工工作量统计
	employeeStats := make(map[string]*EmployeeWorkload)
	
	standardWeeklyHours := 40.0
	
	for _, a := range assignments {
		// 计算工时
		hours := a.EndTime.Sub(a.StartTime).Hours()
		summary.TotalHours += hours
		summary.TotalShifts++

		empID := a.EmployeeID.String()
		// 员工统计
		ew, exists := employeeStats[empID]
		if !exists {
			name := empID
			if emp, ok := employeeMap[empID]; ok {
				name = emp.Name
			}
			ew = &EmployeeWorkload{
				EmployeeID:   empID,
				EmployeeName: name,
			}
			employeeStats[empID] = ew
		}
		ew.TotalHours += hours
		ew.ShiftCount++

		// 日期统计
		daily, exists := summary.ByDate[a.Date]
		if !exists {
			daily = DailyWorkload{Date: a.Date}
		}
		daily.TotalHours += hours
		daily.ShiftCount++
		daily.StaffCount++
		summary.ByDate[a.Date] = daily

		// 班次类型统计
		shiftType := classifyShiftType(a.StartTime)
		summary.ByShiftType[shiftType] += hours
	}

	// 计算加班和利用率
	summary.EmployeeCount = len(employeeStats)
	
	// 计算周数
	weeks := 1.0
	if startDate != "" && endDate != "" {
		start, err1 := time.Parse("2006-01-02", startDate)
		end, err2 := time.Parse("2006-01-02", endDate)
		if err1 == nil && err2 == nil {
			days := end.Sub(start).Hours() / 24
			weeks = days / 7
			if weeks < 1 {
				weeks = 1
			}
		}
	}

	expectedHours := standardWeeklyHours * weeks

	for _, ew := range employeeStats {
		if ew.TotalHours > expectedHours {
			ew.OvertimeHours = ew.TotalHours - expectedHours
			summary.OvertimeHours += ew.OvertimeHours
		}
		ew.Utilization = ew.TotalHours / expectedHours * 100
		summary.ByEmployee = append(summary.ByEmployee, *ew)
	}

	// 计算人均工时
	if summary.EmployeeCount > 0 {
		summary.AvgHoursPerPerson = summary.TotalHours / float64(summary.EmployeeCount)
	}

	return summary
}

// classifyShiftType 分类班次类型
func classifyShiftType(start time.Time) string {
	hour := start.Hour()
	if hour >= 6 && hour < 14 {
		return "morning"
	} else if hour >= 14 && hour < 22 {
		return "afternoon"
	}
	return "night"
}

// sendJSONError 发送JSON错误响应
func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// convertToAssignmentInfo 转换Assignment为stats包类型
func convertToAssignmentInfo(assignments []*model.Assignment) []*stats.AssignmentInfo {
	result := make([]*stats.AssignmentInfo, len(assignments))
	for i, a := range assignments {
		result[i] = &stats.AssignmentInfo{
			ShiftID:      a.ShiftID.String(),
			EmployeeID:   a.EmployeeID.String(),
			EmployeeName: "", // 由统计包从员工列表获取
			Date:         a.Date,
			StartTime:    a.StartTime,
			EndTime:      a.EndTime,
		}
	}
	return result
}

// convertToEmployeeInfo 转换Employee为stats包类型
func convertToEmployeeInfo(employees []*model.Employee) []*stats.EmployeeInfo {
	result := make([]*stats.EmployeeInfo, len(employees))
	for i, e := range employees {
		result[i] = &stats.EmployeeInfo{
			ID:   e.ID.String(),
			Name: e.Name,
		}
	}
	return result
}

// convertToShiftInfo 转换Shift为stats包类型
func convertToShiftInfo(shifts []*model.Shift) []*stats.ShiftInfo {
	result := make([]*stats.ShiftInfo, len(shifts))
	for i, s := range shifts {
		// 解析时间字符串
		start, _ := time.Parse("15:04", s.StartTime)
		end, _ := time.Parse("15:04", s.EndTime)
		
		result[i] = &stats.ShiftInfo{
			ID:             s.ID.String(),
			Date:           "", // Shift模型没有Date字段
			StartTime:      start,
			EndTime:        end,
			Type:           s.ShiftType,
			Position:       "",
			RequiredSkills: nil,
		}
	}
	return result
}

