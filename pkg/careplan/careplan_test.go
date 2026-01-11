package careplan

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

func TestPlanManager_CreatePlan(t *testing.T) {
	manager := NewPlanManager()

	tests := []struct {
		name      string
		level     int
		wantError bool
	}{
		{"等级1", 1, false},
		{"等级3", 3, false},
		{"等级6", 6, false},
		{"无效等级0", 0, true},
		{"无效等级7", 7, true},
	}

	customerID := uuid.New()
	startDate := "2026-01-11"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := manager.CreatePlan(customerID, tt.level, startDate)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if plan.CustomerID != customerID {
					t.Error("CustomerID mismatch")
				}
				if plan.Level != tt.level {
					t.Errorf("Level mismatch: got %d, want %d", plan.Level, tt.level)
				}
			}
		})
	}
}

func TestPlanManager_ValidatePlan(t *testing.T) {
	manager := NewPlanManager()

	tests := []struct {
		name   string
		plan   *model.CarePlan
		hasErr bool
	}{
		{
			name: "有效计划",
			plan: &model.CarePlan{
				Level:       3,
				StartDate:   "2026-01-11",
				WeeklyHours: 10,
				ServiceItems: []model.CareItem{
					{Code: "care", Name: "护理", Duration: 60, Frequency: 5},
				},
			},
			hasErr: false,
		},
		{
			name: "无效等级",
			plan: &model.CarePlan{
				Level:       7,
				StartDate:   "2026-01-11",
				WeeklyHours: 10,
				ServiceItems: []model.CareItem{
					{Code: "care", Name: "护理", Duration: 60, Frequency: 5},
				},
			},
			hasErr: true,
		},
		{
			name: "无开始日期",
			plan: &model.CarePlan{
				Level:       3,
				WeeklyHours: 10,
				ServiceItems: []model.CareItem{
					{Code: "care", Name: "护理", Duration: 60, Frequency: 5},
				},
			},
			hasErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := manager.ValidatePlan(tt.plan)
			hasErrors := len(errs) > 0
			if hasErrors != tt.hasErr {
				t.Errorf("ValidatePlan() hasErrors = %v, expected %v, errs: %v", hasErrors, tt.hasErr, errs)
			}
		})
	}
}

func TestPlanManager_GenerateServiceOrders(t *testing.T) {
	manager := NewPlanManager()

	plan := &model.CarePlan{
		BaseModel:   model.NewBaseModel(),
		CustomerID:  uuid.New(),
		PlanNo:      "CP001",
		Level:       3,
		StartDate:   "2026-01-11",
		WeeklyHours: 6,
		ServiceItems: []model.CareItem{
			{Code: "basic_care", Name: "基础护理", Duration: 60, Frequency: 3},
		},
		Frequency: "3_times_per_week",
		Status:    "active",
	}

	customer := &model.Customer{
		BaseModel: model.NewBaseModel(),
		Name:      "测试客户",
		Address:   "测试地址",
		Location:  &model.Location{Latitude: 39.9, Longitude: 116.4},
	}
	customer.ID = plan.CustomerID

	orders, err := manager.GenerateServiceOrders(plan, customer, "2026-01-13", "2026-01-19")
	if err != nil {
		t.Fatalf("GenerateServiceOrders failed: %v", err)
	}

	if len(orders) == 0 {
		t.Error("Should generate some orders")
	}
}

func TestPlanManager_GetRecommendedCarers(t *testing.T) {
	manager := NewPlanManager()

	plan := &model.CarePlan{
		Level:     3,
		StartDate: "2026-01-11",
		ServiceItems: []model.CareItem{
			{Code: "basic_care", Name: "基础护理", Duration: 60, Frequency: 5},
		},
	}

	caregivers := []*model.Employee{
		{
			BaseModel:      model.NewBaseModel(),
			Name:           "护理员1",
			Skills:         []string{"护理员证", "基础护理", "健康证"},
			Certifications: []string{"护理员证"}, // 必须有这个证书
			Status:         "active",
		},
	}

	recommendations := manager.GetRecommendedCarers(plan, caregivers)

	if len(recommendations) == 0 {
		t.Error("Should have some recommendations")
	}
}

func TestPlanManager_GetRecommendedCarers_NoCert(t *testing.T) {
	manager := NewPlanManager()

	plan := &model.CarePlan{Level: 3, StartDate: "2026-01-11"}

	// 无护理员证的不应该被推荐
	caregivers := []*model.Employee{
		{
			BaseModel: model.NewBaseModel(),
			Name:      "无证人员",
			Skills:    []string{"基础护理"},
			Status:    "active",
		},
	}

	recommendations := manager.GetRecommendedCarers(plan, caregivers)

	if len(recommendations) != 0 {
		t.Error("无护理员证不应被推荐")
	}
}
