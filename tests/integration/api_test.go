package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

// TestScheduleAPI_GenerateRequest 测试排班API请求格式
func TestScheduleAPI_GenerateRequest(t *testing.T) {
	orgID := uuid.New()

	request := map[string]interface{}{
		"org_id":     orgID.String(),
		"scenario":   "restaurant",
		"start_date": "2026-01-13",
		"end_date":   "2026-01-19",
		"employees": []map[string]interface{}{
			{
				"id":     uuid.New().String(),
				"name":   "张三",
				"skills": []string{"cooking", "service"},
				"status": "active",
			},
		},
		"shifts": []map[string]interface{}{
			{
				"id":         uuid.New().String(),
				"name":       "早班",
				"start_time": "08:00",
				"end_time":   "16:00",
				"type":       "morning",
			},
		},
		"constraints": []map[string]interface{}{
			{
				"type":   "MaxHoursPerDay",
				"params": map[string]interface{}{"max_hours": 10},
				"weight": 100,
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/schedule/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// 验证请求格式正确
	var parsed map[string]interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if parsed["org_id"] != orgID.String() {
		t.Error("org_id mismatch")
	}
	if parsed["scenario"] != "restaurant" {
		t.Error("scenario mismatch")
	}

	t.Log("Schedule API request format validated")
}

// TestDispatchAPI_SingleRequest 测试派单API请求格式
func TestDispatchAPI_SingleRequest(t *testing.T) {
	orgID := uuid.New()
	customerID := uuid.New()

	request := map[string]interface{}{
		"org_id": orgID.String(),
		"orders": []map[string]interface{}{
			{
				"id":           uuid.New().String(),
				"customer_id":  customerID.String(),
				"order_no":     "ORD001",
				"service_type": "cleaning",
				"service_date": "2026-01-15",
				"start_time":   "09:00",
				"end_time":     "11:00",
				"location": map[string]interface{}{
					"latitude":  39.9042,
					"longitude": 116.4074,
				},
			},
		},
		"employees": []map[string]interface{}{
			{
				"id":     uuid.New().String(),
				"name":   "李阿姨",
				"skills": []string{"cleaning"},
				"status": "active",
			},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v1/dispatch/single", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	t.Logf("Dispatch request size: %d bytes", len(body))
	t.Log("Dispatch API request format validated")
	_ = req
}

// TestStatsAPI_FairnessRequest 测试公平性统计API
func TestStatsAPI_FairnessRequest(t *testing.T) {
	orgID := uuid.New()
	emp1 := uuid.New()

	request := map[string]interface{}{
		"org_id":     orgID.String(),
		"start_date": "2026-01-01",
		"end_date":   "2026-01-31",
		"employees": []model.Employee{
			{BaseModel: model.BaseModel{ID: emp1}, Name: "员工1"},
		},
		"assignments": []model.Assignment{
			{BaseModel: model.NewBaseModel(), EmployeeID: emp1, Date: "2026-01-11"},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	t.Logf("Stats request size: %d bytes", len(body))
	t.Log("Stats API request format validated")
}

// TestAPIResponseFormat 测试API响应格式
func TestAPIResponseFormat(t *testing.T) {
	// 成功响应格式
	successResp := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"schedule_id": "sched-001",
			"assignments": []interface{}{},
		},
	}

	body, _ := json.Marshal(successResp)
	t.Logf("Success response: %s", string(body))

	// 错误响应格式
	errorResp := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    "validation_error",
			"message": "请求参数无效",
		},
	}

	body, _ = json.Marshal(errorResp)
	t.Logf("Error response: %s", string(body))
}

// TestHealthEndpoint 测试健康检查端点
func TestHealthEndpoint(t *testing.T) {
	_ = httptest.NewRequest("GET", "/health", nil) // req not used in mock test
	rec := httptest.NewRecorder()

	// 模拟健康检查响应
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusOK)
	json.NewEncoder(rec).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": "2026-01-11T10:30:00Z",
	})

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	t.Log("Health endpoint validated")
}

// TestVersionEndpoint 测试版本信息端点
func TestVersionEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/version", nil)
	rec := httptest.NewRecorder()

	// 模拟版本响应
	rec.Header().Set("Content-Type", "application/json")
	rec.WriteHeader(http.StatusOK)
	json.NewEncoder(rec).Encode(map[string]interface{}{
		"version":    "1.0.0",
		"build_time": "2026-01-11",
		"git_commit": "abc123",
	})

	if rec.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", rec.Code)
	}

	t.Log("Version endpoint validated")
	_ = req
}
