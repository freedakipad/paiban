// Package scenario 提供场景测试
package scenario

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/dispatcher"
	"github.com/paiban/paiban/pkg/model"
)

// TestHousekeepingDispatch 测试家政服务派单
func TestHousekeepingDispatch(t *testing.T) {
	engine := dispatcher.NewDispatchEngine()

	// 创建测试订单
	order := &model.ServiceOrder{
		OrderNo:     "HSK202601110001",
		ServiceType: "cleaning",
		ServiceDate: "2026-01-11",
		StartTime:   "09:00",
		EndTime:     "12:00",
		Duration:    180,
		Status:      "pending",
		Skills:      []string{"保洁"},
	}

	// 创建候选员工
	candidates := []*model.Employee{
		{
			BaseModel:      model.BaseModel{ID: uuid.New()},
			Name:           "张阿姨",
			Skills:         []string{"保洁", "收纳"},
			Certifications: []string{"无犯罪证明", "健康证"},
			Status:         "active",
		},
		{
			BaseModel:      model.BaseModel{ID: uuid.New()},
			Name:           "李阿姨",
			Skills:         []string{"烹饪"},
			Certifications: []string{"无犯罪证明", "健康证"},
			Status:         "active",
		},
		{
			BaseModel:      model.BaseModel{ID: uuid.New()},
			Name:           "王阿姨",
			Skills:         []string{"保洁"},
			Certifications: []string{"健康证"}, // 缺少无犯罪证明
			Status:         "active",
		},
	}

	req := &dispatcher.DispatchRequest{
		Order:      order,
		Candidates: candidates,
		MaxResults: 3,
	}

	resp := engine.Dispatch(req)

	t.Logf("派单结果: success=%v", resp.Success)
	if resp.BestMatch != nil {
		t.Logf("最佳匹配: %s, 分数=%.2f", resp.BestMatch.Employee.Name, resp.BestMatch.Score)
	}

	if !resp.Success {
		t.Errorf("派单应该成功")
	}

	if resp.BestMatch == nil {
		t.Fatalf("应该有最佳匹配")
	}

	// 验证最佳匹配是张阿姨（有保洁技能和无犯罪证明）
	if resp.BestMatch.Employee.Name != "张阿姨" {
		t.Errorf("期望最佳匹配是张阿姨，实际是 %s", resp.BestMatch.Employee.Name)
	}
}

// TestHousekeepingCustomerPreference 测试客户偏好
func TestHousekeepingCustomerPreference(t *testing.T) {
	engine := dispatcher.NewDispatchEngine()

	preferredID := uuid.New()
	blockedID := uuid.New()

	// 创建客户（有偏好员工）
	customer := &model.Customer{
		Name:            "王先生",
		PreferredEmpIDs: []uuid.UUID{preferredID},
		BlockedEmpIDs:   []uuid.UUID{blockedID},
	}

	order := &model.ServiceOrder{
		OrderNo:     "HSK202601110002",
		ServiceType: "cleaning",
		ServiceDate: "2026-01-11",
		StartTime:   "14:00",
		EndTime:     "17:00",
		Status:      "pending",
	}

	candidates := []*model.Employee{
		{
			BaseModel:      model.BaseModel{ID: preferredID},
			Name:           "偏好员工",
			Skills:         []string{"保洁"},
			Certifications: []string{"无犯罪证明"},
			Status:         "active",
		},
		{
			BaseModel:      model.BaseModel{ID: blockedID},
			Name:           "黑名单员工",
			Skills:         []string{"保洁"},
			Certifications: []string{"无犯罪证明"},
			Status:         "active",
		},
		{
			BaseModel:      model.BaseModel{ID: uuid.New()},
			Name:           "普通员工",
			Skills:         []string{"保洁"},
			Certifications: []string{"无犯罪证明"},
			Status:         "active",
		},
	}

	req := &dispatcher.DispatchRequest{
		Order:      order,
		Candidates: candidates,
		Customer:   customer,
		MaxResults: 3,
	}

	resp := engine.Dispatch(req)

	t.Logf("派单结果: success=%v", resp.Success)

	if !resp.Success || resp.BestMatch == nil {
		t.Fatalf("派单应该成功且有最佳匹配")
	}

	// 最佳匹配应该是偏好员工
	if resp.BestMatch.Employee.Name != "偏好员工" {
		t.Errorf("期望最佳匹配是偏好员工，实际是 %s", resp.BestMatch.Employee.Name)
	}

	// 黑名单员工不应该出现在可行解中
	for _, alt := range resp.Alternatives {
		if alt.Employee.Name == "黑名单员工" && alt.Feasible {
			t.Errorf("黑名单员工不应该是可行解")
		}
	}
}

// TestHousekeepingBatchDispatch 测试批量派单
func TestHousekeepingBatchDispatch(t *testing.T) {
	engine := dispatcher.NewDispatchEngine()

	// 创建多个订单
	orders := []*model.ServiceOrder{
		{
			OrderNo:     "HSK202601110003",
			ServiceType: "cleaning",
			ServiceDate: "2026-01-11",
			StartTime:   "09:00",
			EndTime:     "11:00",
			Status:      "pending",
		},
		{
			OrderNo:     "HSK202601110004",
			ServiceType: "cleaning",
			ServiceDate: "2026-01-11",
			StartTime:   "11:30",
			EndTime:     "13:30",
			Status:      "pending",
		},
		{
			OrderNo:     "HSK202601110005",
			ServiceType: "cooking",
			ServiceDate: "2026-01-11",
			StartTime:   "17:00",
			EndTime:     "19:00",
			Status:      "pending",
		},
	}

	candidates := []*model.Employee{
		{
			BaseModel:      model.BaseModel{ID: uuid.New()},
			Name:           "张阿姨",
			Skills:         []string{"保洁", "烹饪"},
			Certifications: []string{"无犯罪证明", "健康证"},
			Status:         "active",
		},
		{
			BaseModel:      model.BaseModel{ID: uuid.New()},
			Name:           "李阿姨",
			Skills:         []string{"保洁"},
			Certifications: []string{"无犯罪证明", "健康证"},
			Status:         "active",
		},
	}

	responses := engine.BatchDispatch(orders, candidates, nil)

	successCount := 0
	for i, resp := range responses {
		t.Logf("订单 %d (%s): success=%v", i+1, orders[i].OrderNo, resp.Success)
		if resp.Success {
			successCount++
			if resp.BestMatch != nil {
				t.Logf("  分配给: %s", resp.BestMatch.Employee.Name)
			}
		}
	}

	t.Logf("总计: %d/%d 成功", successCount, len(orders))

	// 至少应该有2个订单成功派单
	if successCount < 2 {
		t.Errorf("期望至少2个订单成功派单，实际 %d", successCount)
	}
}

// TestHousekeepingTimeConflict 测试时间冲突检测
func TestHousekeepingTimeConflict(t *testing.T) {
	engine := dispatcher.NewDispatchEngine()

	emp := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "张阿姨",
		Skills:         []string{"保洁"},
		Certifications: []string{"无犯罪证明"},
		Status:         "active",
	}

	// 已有订单
	existingOrder := &model.ServiceOrder{
		OrderNo:     "HSK202601110010",
		ServiceType: "cleaning",
		ServiceDate: "2026-01-11",
		StartTime:   "09:00",
		EndTime:     "11:00",
		EmployeeID:  &emp.ID,
		Status:      "assigned",
	}

	// 新订单与已有订单时间重叠
	newOrder := &model.ServiceOrder{
		OrderNo:     "HSK202601110011",
		ServiceType: "cleaning",
		ServiceDate: "2026-01-11",
		StartTime:   "10:00",
		EndTime:     "12:00",
		Status:      "pending",
	}

	req := &dispatcher.DispatchRequest{
		Order:       newOrder,
		Candidates:  []*model.Employee{emp},
		TodayOrders: []*model.ServiceOrder{existingOrder},
		MaxResults:  1,
	}

	resp := engine.Dispatch(req)

	t.Logf("时间冲突测试: success=%v", resp.Success)

	// 应该无法派单（时间冲突）
	if resp.Success && resp.BestMatch != nil && resp.BestMatch.Feasible {
		t.Errorf("应该检测到时间冲突，派单失败")
	}
}
