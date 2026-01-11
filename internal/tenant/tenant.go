// Package tenant 提供多租户支持
package tenant

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTenantNotFound = errors.New("租户不存在")
	ErrInvalidTenant  = errors.New("无效的租户")
	ErrTenantDisabled = errors.New("租户已禁用")
)

// Tenant 租户
type Tenant struct {
	ID          uuid.UUID     `json:"id"`
	Code        string        `json:"code"`         // 租户编码
	Name        string        `json:"name"`         // 租户名称
	Type        string        `json:"type"`         // enterprise/individual
	Status      string        `json:"status"`       // active/suspended/expired
	Settings    TenantSettings `json:"settings"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiredAt   *time.Time    `json:"expired_at,omitempty"`
}

// TenantSettings 租户配置
type TenantSettings struct {
	MaxEmployees    int      `json:"max_employees"`     // 最大员工数
	MaxOrders       int      `json:"max_orders_per_day"` // 每日最大订单数
	AllowedScenarios []string `json:"allowed_scenarios"` // 允许的场景
	Features        []string `json:"features"`          // 启用的功能
	APIRateLimit    int      `json:"api_rate_limit"`    // API速率限制
	DataRetention   int      `json:"data_retention_days"` // 数据保留天数
}

// IsActive 检查租户是否活跃
func (t *Tenant) IsActive() bool {
	if t.Status != "active" {
		return false
	}
	if t.ExpiredAt != nil && t.ExpiredAt.Before(time.Now()) {
		return false
	}
	return true
}

// HasFeature 检查租户是否拥有某功能
func (t *Tenant) HasFeature(feature string) bool {
	for _, f := range t.Settings.Features {
		if f == feature || f == "*" {
			return true
		}
	}
	return false
}

// HasScenario 检查租户是否允许某场景
func (t *Tenant) HasScenario(scenario string) bool {
	for _, s := range t.Settings.AllowedScenarios {
		if s == scenario || s == "*" {
			return true
		}
	}
	return false
}

// TenantManager 租户管理器
type TenantManager struct {
	tenants map[string]*Tenant // code -> tenant
	mu      sync.RWMutex
}

// NewTenantManager 创建租户管理器
func NewTenantManager() *TenantManager {
	return &TenantManager{
		tenants: make(map[string]*Tenant),
	}
}

// Register 注册租户
func (m *TenantManager) Register(tenant *Tenant) error {
	if tenant == nil || tenant.Code == "" {
		return ErrInvalidTenant
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.tenants[tenant.Code] = tenant
	return nil
}

// Get 获取租户
func (m *TenantManager) Get(code string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, exists := m.tenants[code]
	if !exists {
		return nil, ErrTenantNotFound
	}

	if !tenant.IsActive() {
		return nil, ErrTenantDisabled
	}

	return tenant, nil
}

// GetByID 通过ID获取租户
func (m *TenantManager) GetByID(id uuid.UUID) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, tenant := range m.tenants {
		if tenant.ID == id {
			if !tenant.IsActive() {
				return nil, ErrTenantDisabled
			}
			return tenant, nil
		}
	}

	return nil, ErrTenantNotFound
}

// List 列出所有租户
func (m *TenantManager) List() []*Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Tenant, 0, len(m.tenants))
	for _, t := range m.tenants {
		result = append(result, t)
	}
	return result
}

// Remove 移除租户
func (m *TenantManager) Remove(code string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tenants, code)
}

// TenantContext 租户上下文键
type tenantContextKey struct{}

// WithTenant 将租户添加到上下文
func WithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

// FromContext 从上下文获取租户
func FromContext(ctx context.Context) (*Tenant, bool) {
	tenant, ok := ctx.Value(tenantContextKey{}).(*Tenant)
	return tenant, ok
}

// DefaultTenantSettings 默认租户配置
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		MaxEmployees:     100,
		MaxOrders:        1000,
		AllowedScenarios: []string{"restaurant", "factory", "housekeeping", "nursing"},
		Features:         []string{"schedule", "dispatch", "stats"},
		APIRateLimit:     100,
		DataRetention:    365,
	}
}

// CreateDefaultTenant 创建默认租户（开发测试用）
func CreateDefaultTenant() *Tenant {
	return &Tenant{
		ID:       uuid.New(),
		Code:     "default",
		Name:     "默认租户",
		Type:     "enterprise",
		Status:   "active",
		Settings: DefaultTenantSettings(),
		CreatedAt: time.Now(),
	}
}

