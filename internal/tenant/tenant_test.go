package tenant

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTenant_IsActive(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		tenant   *Tenant
		expected bool
	}{
		{
			name:     "活跃租户",
			tenant:   &Tenant{Status: "active"},
			expected: true,
		},
		{
			name:     "暂停租户",
			tenant:   &Tenant{Status: "suspended"},
			expected: false,
		},
		{
			name:     "未过期租户",
			tenant:   &Tenant{Status: "active", ExpiredAt: &future},
			expected: true,
		},
		{
			name:     "已过期租户",
			tenant:   &Tenant{Status: "active", ExpiredAt: &past},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := tt.tenant.IsActive(); result != tt.expected {
				t.Errorf("IsActive() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTenant_HasFeature(t *testing.T) {
	tenant := &Tenant{
		Settings: TenantSettings{
			Features: []string{"schedule", "dispatch"},
		},
	}

	if !tenant.HasFeature("schedule") {
		t.Error("应有schedule功能")
	}
	if !tenant.HasFeature("dispatch") {
		t.Error("应有dispatch功能")
	}
	if tenant.HasFeature("stats") {
		t.Error("不应有stats功能")
	}

	// 测试通配符
	tenant2 := &Tenant{
		Settings: TenantSettings{
			Features: []string{"*"},
		},
	}
	if !tenant2.HasFeature("anything") {
		t.Error("通配符应匹配任何功能")
	}
}

func TestTenant_HasScenario(t *testing.T) {
	tenant := &Tenant{
		Settings: TenantSettings{
			AllowedScenarios: []string{"restaurant", "factory"},
		},
	}

	if !tenant.HasScenario("restaurant") {
		t.Error("应有restaurant场景")
	}
	if tenant.HasScenario("nursing") {
		t.Error("不应有nursing场景")
	}
}

func TestTenantManager_RegisterAndGet(t *testing.T) {
	manager := NewTenantManager()

	tenant := &Tenant{
		ID:     uuid.New(),
		Code:   "test",
		Name:   "测试租户",
		Status: "active",
	}

	// 注册
	err := manager.Register(tenant)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}

	// 获取
	got, err := manager.Get("test")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if got.Code != "test" {
		t.Errorf("Got wrong tenant: %v", got)
	}

	// 获取不存在的
	_, err = manager.Get("nonexistent")
	if err != ErrTenantNotFound {
		t.Errorf("Expected ErrTenantNotFound, got: %v", err)
	}
}

func TestTenantManager_GetByID(t *testing.T) {
	manager := NewTenantManager()
	id := uuid.New()

	tenant := &Tenant{
		ID:     id,
		Code:   "test",
		Status: "active",
	}
	manager.Register(tenant)

	got, err := manager.GetByID(id)
	if err != nil {
		t.Errorf("GetByID failed: %v", err)
	}
	if got.ID != id {
		t.Errorf("Got wrong tenant")
	}
}

func TestTenantContext(t *testing.T) {
	tenant := &Tenant{Code: "test"}
	ctx := WithTenant(context.Background(), tenant)

	got, ok := FromContext(ctx)
	if !ok {
		t.Error("FromContext should return true")
	}
	if got.Code != "test" {
		t.Error("Got wrong tenant from context")
	}

	// 空上下文
	_, ok = FromContext(context.Background())
	if ok {
		t.Error("Empty context should return false")
	}
}

func TestDefaultTenantSettings(t *testing.T) {
	settings := DefaultTenantSettings()

	if settings.MaxEmployees != 100 {
		t.Errorf("Expected MaxEmployees=100, got %d", settings.MaxEmployees)
	}
	if len(settings.AllowedScenarios) != 4 {
		t.Errorf("Expected 4 scenarios, got %d", len(settings.AllowedScenarios))
	}
}

func TestCreateDefaultTenant(t *testing.T) {
	tenant := CreateDefaultTenant()

	if tenant.Code != "default" {
		t.Errorf("Expected code='default', got %s", tenant.Code)
	}
	if tenant.Status != "active" {
		t.Errorf("Expected status='active', got %s", tenant.Status)
	}
}

