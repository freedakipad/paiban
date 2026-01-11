// Package constraint 定义约束接口和管理器
package constraint

import (
	"fmt"
	"sort"
	"sync"

	"github.com/paiban/paiban/pkg/logger"
	"github.com/paiban/paiban/pkg/model"
)

// Manager 约束管理器
type Manager struct {
	constraints []Constraint
	mu          sync.RWMutex
	logger      *logger.SchedulerLogger
}

// NewManager 创建约束管理器
func NewManager() *Manager {
	return &Manager{
		constraints: make([]Constraint, 0),
		logger:      logger.NewSchedulerLogger(),
	}
}

// Register 注册约束
func (m *Manager) Register(c Constraint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在同类型约束
	for i, existing := range m.constraints {
		if existing.Type() == c.Type() {
			m.constraints[i] = c // 替换
			return
		}
	}

	m.constraints = append(m.constraints, c)

	// 按类别和权重排序：硬约束在前，权重高的在前
	sort.Slice(m.constraints, func(i, j int) bool {
		ci, cj := m.constraints[i], m.constraints[j]
		if ci.Category() != cj.Category() {
			return ci.Category() == CategoryHard
		}
		return ci.Weight() > cj.Weight()
	})
}

// Unregister 注销约束
func (m *Manager) Unregister(t Type) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, c := range m.constraints {
		if c.Type() == t {
			m.constraints = append(m.constraints[:i], m.constraints[i+1:]...)
			return
		}
	}
}

// GetConstraint 获取约束
func (m *Manager) GetConstraint(t Type) Constraint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.constraints {
		if c.Type() == t {
			return c
		}
	}
	return nil
}

// GetAll 获取所有约束
func (m *Manager) GetAll() []Constraint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]Constraint, len(m.constraints))
	copy(result, m.constraints)
	return result
}

// GetByCategory 按类别获取约束
func (m *Manager) GetByCategory(cat Category) []Constraint {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Constraint
	for _, c := range m.constraints {
		if c.Category() == cat {
			result = append(result, c)
		}
	}
	return result
}

// Evaluate 评估所有约束
func (m *Manager) Evaluate(ctx *Context) *Result {
	m.mu.RLock()
	constraints := make([]Constraint, len(m.constraints))
	copy(constraints, m.constraints)
	m.mu.RUnlock()

	result := &Result{
		IsValid:        true,
		TotalPenalty:   0,
		HardViolations: make([]ViolationDetail, 0),
		SoftViolations: make([]ViolationDetail, 0),
	}

	maxPenalty := 0

	for _, c := range constraints {
		valid, penalty, details := c.Evaluate(ctx)

		// 累加最大可能惩罚值（用于计算得分）
		maxPenalty += c.Weight() * 100 // 假设每个约束最多违反100次

		if !valid {
			result.TotalPenalty += penalty

			for _, d := range details {
				if c.Category() == CategoryHard {
					result.IsValid = false
					result.HardViolations = append(result.HardViolations, d)
					m.logger.ConstraintViolation(c.Name(), d.Message)
				} else {
					result.SoftViolations = append(result.SoftViolations, d)
				}
			}
		}
	}

	result.CalculateScore(maxPenalty)
	return result
}

// EvaluateAssignment 评估单个分配
func (m *Manager) EvaluateAssignment(ctx *Context, assignment *model.Assignment) (bool, int, []ViolationDetail) {
	m.mu.RLock()
	constraints := make([]Constraint, len(m.constraints))
	copy(constraints, m.constraints)
	m.mu.RUnlock()

	var violations []ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, c := range constraints {
		valid, penalty := c.EvaluateAssignment(ctx, assignment)
		if !valid {
			totalPenalty += penalty
			violations = append(violations, ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     assignment.EmployeeID,
				Date:           assignment.Date,
				Message:        fmt.Sprintf("违反约束: %s", c.Name()),
				Severity:       string(c.Category()),
				Penalty:        penalty,
			})

			if c.Category() == CategoryHard {
				isValid = false
			}
		}
	}

	return isValid, totalPenalty, violations
}

// CanAssign 检查是否可以进行某个分配
func (m *Manager) CanAssign(ctx *Context, assignment *model.Assignment) (bool, string) {
	// 只检查硬约束
	hardConstraints := m.GetByCategory(CategoryHard)

	for _, c := range hardConstraints {
		valid, _ := c.EvaluateAssignment(ctx, assignment)
		if !valid {
			return false, fmt.Sprintf("违反硬约束: %s", c.Name())
		}
	}

	return true, ""
}

// GetPenalty 计算分配的惩罚值
func (m *Manager) GetPenalty(ctx *Context, assignment *model.Assignment) int {
	_, penalty, _ := m.EvaluateAssignment(ctx, assignment)
	return penalty
}

// Clear 清除所有约束
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.constraints = make([]Constraint, 0)
}

// Count 返回约束数量
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.constraints)
}

// Summary 返回约束摘要
func (m *Manager) Summary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hard := 0
	soft := 0
	for _, c := range m.constraints {
		if c.Category() == CategoryHard {
			hard++
		} else {
			soft++
		}
	}

	return map[string]interface{}{
		"total": len(m.constraints),
		"hard":  hard,
		"soft":  soft,
	}
}
