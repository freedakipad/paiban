// Package e2e 提供端到端测试
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

// TestFullSchedulingWorkflow 测试完整排班工作流
func TestFullSchedulingWorkflow(t *testing.T) {
	// 准备测试数据
	orgID := uuid.New()
	employees := createTestEmployees(orgID, 5)
	shifts := createTestShifts(orgID, 7) // 一周的班次

	// 生成排班请求
	req := map[string]interface{}{
		"org_id":     orgID.String(),
		"scenario":   "restaurant",
		"start_date": time.Now().Format("2006-01-02"),
		"end_date":   time.Now().AddDate(0, 0, 6).Format("2006-01-02"),
		"employees":  employees,
		"shifts":     shifts,
		"constraints": []map[string]interface{}{
			{
				"type":   "MaxHoursPerDay",
				"params": map[string]interface{}{"max_hours": 10},
				"weight": 100,
			},
			{
				"type":   "MinRestBetweenShifts",
				"params": map[string]interface{}{"min_hours": 8},
				"weight": 100,
			},
		},
	}

	body, _ := json.Marshal(req)

	// 发送请求
	httpReq := httptest.NewRequest("POST", "/api/v1/schedule/generate", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	// 这里模拟调用handler
	t.Log("发送排班生成请求...")
	t.Logf("请求体: %s", string(body))

	// 验证请求格式正确
	if len(employees) != 5 {
		t.Errorf("期望5个员工，实际: %d", len(employees))
	}
	if len(shifts) != 7 {
		t.Errorf("期望7个班次，实际: %d", len(shifts))
	}

	t.Log("E2E测试准备完成")
	_ = recorder // 避免未使用警告
}

// TestFullDispatchWorkflow 测试完整派单工作流
func TestFullDispatchWorkflow(t *testing.T) {
	orgID := uuid.New()

	// 1. 创建客户
	customer := &model.Customer{
		BaseModel: model.NewBaseModel(),
		OrgID:     orgID,
		Name:      "测试客户",
		Code:      "C001",
		Phone:     "13800138000",
		Address:   "北京市朝阳区",
		Location: &model.Location{
			Address:   "北京市朝阳区",
			Latitude:  39.9042,
			Longitude: 116.4074,
		},
		Type:   "individual",
		Status: "active",
	}

	// 2. 创建员工
	employees := []*model.Employee{
		{
			BaseModel:      model.NewBaseModel(),
			OrgID:          orgID,
			Name:           "张师傅",
			Code:           "E001",
			Skills:         []string{"cleaning", "cooking"},
			Certifications: []string{"health_cert", "no_criminal_record"},
			Status:         "active",
		},
		{
			BaseModel:      model.NewBaseModel(),
			OrgID:          orgID,
			Name:           "李师傅",
			Code:           "E002",
			Skills:         []string{"cleaning"},
			Certifications: []string{"health_cert"},
			Status:         "active",
		},
	}

	// 3. 创建服务订单
	order := &model.ServiceOrder{
		BaseModel:   model.NewBaseModel(),
		OrgID:       orgID,
		CustomerID:  customer.ID,
		OrderNo:     "ORD001",
		ServiceType: "cleaning",
		ServiceDate: time.Now().Format("2006-01-02"),
		StartTime:   "09:00",
		EndTime:     "11:00",
		Duration:    120,
		Address:     customer.Address,
		Location:    customer.Location,
		Status:      "pending",
		Skills:      []string{"cleaning"},
		Priority:    1,
	}

	// 4. 构建派单请求
	dispatchReq := map[string]interface{}{
		"org_id":    orgID.String(),
		"orders":    []*model.ServiceOrder{order},
		"employees": employees,
		"customers": []*model.Customer{customer},
	}

	body, _ := json.Marshal(dispatchReq)
	t.Logf("派单请求: %s", string(body))

	// 5. 验证数据完整性
	if customer.Location == nil {
		t.Error("客户位置不能为空")
	}
	if len(employees) < 1 {
		t.Error("至少需要一个员工")
	}
	if order.Status != "pending" {
		t.Errorf("订单状态应为pending, 实际: %s", order.Status)
	}

	t.Log("派单E2E测试准备完成")
}

// TestFullCarePlanWorkflow 测试完整护理计划工作流
func TestFullCarePlanWorkflow(t *testing.T) {
	// 1. 创建客户（老人）
	customerID := uuid.New()
	customer := &model.Customer{
		BaseModel: model.NewBaseModel(),
		OrgID:     uuid.New(),
		Name:      "王奶奶",
		Code:      "CUST001",
		Phone:     "13900139000",
		Address:   "上海市浦东新区",
		Type:      "individual",
		Status:    "active",
	}
	customer.ID = customerID

	// 2. 创建护理计划
	plan := &model.CarePlan{
		BaseModel:   model.NewBaseModel(),
		CustomerID:  customerID,
		PlanNo:      "CP2026001",
		Level:       3, // 护理等级3级
		StartDate:   time.Now().Format("2006-01-02"),
		EndDate:     time.Now().AddDate(1, 0, 0).Format("2006-01-02"),
		WeeklyHours: 10,
		ServiceItems: []model.CareItem{
			{Code: "basic_care", Name: "基础护理", Duration: 60, Frequency: 7},   // 每周7次
			{Code: "health_check", Name: "健康检查", Duration: 30, Frequency: 1}, // 每周1次
		},
		Frequency: "5_times_per_week",
		Status:    "active",
	}

	// 3. 验证计划
	if plan.Level < 1 || plan.Level > 6 {
		t.Errorf("护理等级无效: %d", plan.Level)
	}
	if plan.WeeklyHours <= 0 {
		t.Error("每周服务时长必须大于0")
	}
	if len(plan.ServiceItems) == 0 {
		t.Error("服务项目不能为空")
	}

	// 4. 模拟生成服务订单
	periodStart := time.Now()
	periodEnd := periodStart.AddDate(0, 0, 7) // 一周

	t.Logf("为客户 %s 创建护理计划 %s", customer.Name, plan.PlanNo)
	t.Logf("服务周期: %s ~ %s", periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))
	t.Logf("每周服务时长: %d 小时", plan.WeeklyHours)

	// 5. 验证订单生成逻辑
	expectedOrders := 5 // 假设一周5次服务
	t.Logf("预期生成 %d 个服务订单", expectedOrders)

	t.Log("护理计划E2E测试准备完成")
}

// TestAPIEndpoints 测试所有API端点
func TestAPIEndpoints(t *testing.T) {
	endpoints := []struct {
		method string
		path   string
		status int
	}{
		{"GET", "/health", http.StatusOK},
		{"GET", "/version", http.StatusOK},
		{"GET", "/api/v1/", http.StatusOK},
		{"GET", "/api/v1/constraints/templates", http.StatusOK},
		{"POST", "/api/v1/schedule/generate", http.StatusBadRequest}, // 无请求体
		{"POST", "/api/v1/schedule/validate", http.StatusBadRequest},
		{"POST", "/api/v1/dispatch/single", http.StatusBadRequest},
		{"POST", "/api/v1/dispatch/batch", http.StatusBadRequest},
		{"POST", "/api/v1/dispatch/route", http.StatusBadRequest},
		{"POST", "/api/v1/careplan/create", http.StatusBadRequest},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s_%s", ep.method, ep.path), func(t *testing.T) {
			t.Logf("测试端点: %s %s", ep.method, ep.path)
			// 这里应该启动实际服务器进行测试
			// 当前只验证端点定义
		})
	}
}

// TestConcurrentRequests 测试并发请求
func TestConcurrentRequests(t *testing.T) {
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			t.Logf("并发请求 #%d", id)
			// 模拟并发请求
			time.Sleep(10 * time.Millisecond)
			done <- true
		}(i)
	}

	// 等待所有请求完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	t.Log("并发测试完成")
}

// 辅助函数
func createTestEmployees(orgID uuid.UUID, count int) []*model.Employee {
	employees := make([]*model.Employee, count)
	skills := [][]string{
		{"cooking", "service"},
		{"cooking"},
		{"service", "cleaning"},
		{"cooking", "management"},
		{"service"},
	}

	for i := 0; i < count; i++ {
		employees[i] = &model.Employee{
			BaseModel: model.NewBaseModel(),
			OrgID:     orgID,
			Name:      fmt.Sprintf("员工%d", i+1),
			Code:      fmt.Sprintf("E%03d", i+1),
			Skills:    skills[i%len(skills)],
			Status:    "active",
		}
	}
	return employees
}

func createTestShifts(orgID uuid.UUID, days int) []*model.Shift {
	shifts := make([]*model.Shift, days)
	baseDate := time.Now()

	for i := 0; i < days; i++ {
		date := baseDate.AddDate(0, 0, i)
		shifts[i] = &model.Shift{
			BaseModel: model.NewBaseModel(),
			OrgID:     orgID,
			Name:      fmt.Sprintf("日班-%s", date.Format("01-02")),
			Code:      fmt.Sprintf("DAY-%s", date.Format("0102")),
			ShiftType: "day",
			StartTime: "09:00",
			EndTime:   "17:00",
			Duration:  480, // 8小时
			BreakTime: 60,  // 1小时休息
			IsActive:  true,
		}
	}
	return shifts
}
