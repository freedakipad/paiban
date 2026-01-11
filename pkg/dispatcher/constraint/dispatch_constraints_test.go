package constraint

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

func TestServiceAreaMatchConstraint_Evaluate(t *testing.T) {
	constraint := NewServiceAreaMatchConstraint(10) // 10公里

	order := &model.ServiceOrder{
		Location: &model.Location{Latitude: 39.91, Longitude: 116.41},
	}
	employee := &model.Employee{
		BaseModel: model.BaseModel{ID: uuid.New()},
	}

	tests := []struct {
		name         string
		empLocation  *model.Location
		expected     bool
	}{
		{
			name:        "在范围内",
			empLocation: &model.Location{Latitude: 39.90, Longitude: 116.40},
			expected:    true,
		},
		{
			name:        "超出范围",
			empLocation: &model.Location{Latitude: 40.0, Longitude: 117.0},
			expected:    false,
		},
		{
			name:        "无位置信息",
			empLocation: nil,
			expected:    true, // 无法判断，默认通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &DispatchContext{
				EmployeeLocation: tt.empLocation,
			}
			passed, _, _ := constraint.Evaluate(order, employee, ctx)
			if passed != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, passed)
			}
		})
	}
}

func TestMaxOrdersPerDayConstraint_Evaluate(t *testing.T) {
	constraint := NewMaxOrdersPerDayConstraint(5)

	order := &model.ServiceOrder{ServiceDate: "2026-01-11"}
	employee := &model.Employee{BaseModel: model.BaseModel{ID: uuid.New()}}

	tests := []struct {
		name       string
		orderCount int
		expected   bool
	}{
		{"未达上限", 3, true},
		{"达到上限", 5, false},
		{"超过上限", 6, false},
		{"无订单", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule := make([]*model.ServiceOrder, tt.orderCount)
			for i := 0; i < tt.orderCount; i++ {
				schedule[i] = &model.ServiceOrder{ServiceDate: "2026-01-11"}
			}

			ctx := &DispatchContext{
				EmployeeOrders: schedule,
			}

			passed, _, _ := constraint.Evaluate(order, employee, ctx)
			if passed != tt.expected {
				t.Errorf("Expected %v for %d orders, got %v", tt.expected, tt.orderCount, passed)
			}
		})
	}
}

func TestCustomerPreferenceConstraint_Evaluate(t *testing.T) {
	constraint := NewCustomerPreferenceConstraint()

	empID := uuid.New()
	order := &model.ServiceOrder{}
	employee := &model.Employee{BaseModel: model.BaseModel{ID: empID}}

	tests := []struct {
		name         string
		preferredIDs []uuid.UUID
		blockedIDs   []uuid.UUID
		expectedPass bool
	}{
		{
			name:         "无偏好",
			preferredIDs: nil,
			blockedIDs:   nil,
			expectedPass: true,
		},
		{
			name:         "偏好员工",
			preferredIDs: []uuid.UUID{empID},
			blockedIDs:   nil,
			expectedPass: true,
		},
		{
			name:         "黑名单员工",
			preferredIDs: nil,
			blockedIDs:   []uuid.UUID{empID},
			expectedPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &DispatchContext{
				Customer: &model.Customer{
					PreferredEmpIDs: tt.preferredIDs,
					BlockedEmpIDs:   tt.blockedIDs,
				},
			}

			passed, _, _ := constraint.Evaluate(order, employee, ctx)
			if passed != tt.expectedPass {
				t.Errorf("Expected pass=%v, got %v", tt.expectedPass, passed)
			}
		})
	}
}

func TestSkillMatchConstraint_Evaluate(t *testing.T) {
	constraint := NewSkillMatchConstraint()

	tests := []struct {
		name        string
		orderSkills []string
		empSkills   []string
		expected    bool
	}{
		{"无技能要求", nil, []string{"cooking"}, true},
		{"技能匹配", []string{"cooking"}, []string{"cooking", "cleaning"}, true},
		{"技能不匹配", []string{"nursing"}, []string{"cooking"}, false},
		{"多技能全匹配", []string{"a", "b"}, []string{"a", "b", "c"}, true},
		{"多技能部分匹配", []string{"a", "b"}, []string{"a"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &model.ServiceOrder{Skills: tt.orderSkills}
			employee := &model.Employee{
				BaseModel: model.BaseModel{ID: uuid.New()},
				Skills:    tt.empSkills,
			}
			ctx := &DispatchContext{}

			passed, _, _ := constraint.Evaluate(order, employee, ctx)
			if passed != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, passed)
			}
		})
	}
}

func TestDefaultDispatchConstraints(t *testing.T) {
	constraints := DefaultDispatchConstraints()

	if len(constraints) == 0 {
		t.Error("Should return default constraints")
	}

	// 检查包含必要的约束
	names := make(map[string]bool)
	for _, c := range constraints {
		names[c.Name()] = true
	}

	required := []string{"ServiceAreaMatch", "MaxOrdersPerDay", "SkillMatch"}
	for _, name := range required {
		if !names[name] {
			t.Errorf("Missing required constraint: %s", name)
		}
	}
}

func TestTravelTimeBufferConstraint_NoExistingOrders(t *testing.T) {
	constraint := NewTravelTimeBufferConstraint(30)

	order := &model.ServiceOrder{
		StartTime: "09:00",
		EndTime:   "11:00",
	}
	employee := &model.Employee{BaseModel: model.BaseModel{ID: uuid.New()}}
	ctx := &DispatchContext{
		EmployeeOrders: nil,
	}

	passed, _, _ := constraint.Evaluate(order, employee, ctx)
	if !passed {
		t.Error("Should pass when no existing orders")
	}
}
