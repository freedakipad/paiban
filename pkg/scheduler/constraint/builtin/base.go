// Package builtin 提供内置约束实现
package builtin

import (
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// BaseConstraint 约束基类
type BaseConstraint struct {
	name     string
	typ      constraint.Type
	category constraint.Category
	weight   int
	config   map[string]interface{}
}

// NewBaseConstraint 创建基础约束
func NewBaseConstraint(name string, typ constraint.Type, cat constraint.Category, weight int) *BaseConstraint {
	return &BaseConstraint{
		name:     name,
		typ:      typ,
		category: cat,
		weight:   weight,
		config:   make(map[string]interface{}),
	}
}

// Name 返回约束名称
func (c *BaseConstraint) Name() string { return c.name }

// Type 返回约束类型
func (c *BaseConstraint) Type() constraint.Type { return c.typ }

// Category 返回约束类别
func (c *BaseConstraint) Category() constraint.Category { return c.category }

// Weight 返回约束权重
func (c *BaseConstraint) Weight() int { return c.weight }

// SetConfig 设置配置
func (c *BaseConstraint) SetConfig(config map[string]interface{}) {
	c.config = config
}

// GetConfig 获取配置
func (c *BaseConstraint) GetConfig() map[string]interface{} {
	return c.config
}

// GetConfigInt 获取整数配置
func (c *BaseConstraint) GetConfigInt(key string, defaultVal int) int {
	if val, ok := c.config[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}
	return defaultVal
}

// GetConfigFloat 获取浮点数配置
func (c *BaseConstraint) GetConfigFloat(key string, defaultVal float64) float64 {
	if val, ok := c.config[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultVal
}

// GetConfigString 获取字符串配置
func (c *BaseConstraint) GetConfigString(key string, defaultVal string) string {
	if val, ok := c.config[key].(string); ok {
		return val
	}
	return defaultVal
}

// GetConfigBool 获取布尔配置
func (c *BaseConstraint) GetConfigBool(key string, defaultVal bool) bool {
	if val, ok := c.config[key].(bool); ok {
		return val
	}
	return defaultVal
}

// CreateViolation 创建违反详情
func (c *BaseConstraint) CreateViolation(empID, date, message string, penalty int) constraint.ViolationDetail {
	severity := "warning"
	if c.category == constraint.CategoryHard {
		severity = "error"
	}

	return constraint.ViolationDetail{
		ConstraintType: c.typ,
		ConstraintName: c.name,
		Date:           date,
		Message:        message,
		Severity:       severity,
		Penalty:        penalty,
	}
}

// Evaluate 默认评估实现（子类需覆盖）
func (c *BaseConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	return true, 0, nil
}

// EvaluateAssignment 默认分配评估实现（子类需覆盖）
func (c *BaseConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

