package constraint

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

func TestManager_Register(t *testing.T) {
	manager := NewManager()

	c := &MockConstraint{
		name:     "test",
		typ:      Type("test_type"),
		category: CategoryHard,
	}
	manager.Register(c)

	constraints := manager.GetAll()
	if len(constraints) != 1 {
		t.Errorf("Expected 1 constraint, got %d", len(constraints))
	}
}

func TestManager_GetByCategory(t *testing.T) {
	manager := NewManager()

	hard := &MockConstraint{name: "hard1", typ: Type("hard1"), category: CategoryHard}
	soft := &MockConstraint{name: "soft1", typ: Type("soft1"), category: CategorySoft}
	manager.Register(hard)
	manager.Register(soft)

	hardConstraints := manager.GetByCategory(CategoryHard)
	if len(hardConstraints) != 1 {
		t.Errorf("Expected 1 hard constraint, got %d", len(hardConstraints))
	}

	softConstraints := manager.GetByCategory(CategorySoft)
	if len(softConstraints) != 1 {
		t.Errorf("Expected 1 soft constraint, got %d", len(softConstraints))
	}
}

func TestManager_Evaluate(t *testing.T) {
	manager := NewManager()

	// 注册一个通过的约束
	pass := &MockConstraint{
		name:     "pass",
		typ:      Type("pass_type"),
		category: CategoryHard,
		pass:     true,
	}
	manager.Register(pass)

	ctx := NewContext(uuid.New(), "2026-01-11", "2026-01-17")

	result := manager.Evaluate(ctx)

	if result.TotalPenalty != 0 {
		t.Errorf("Expected 0 penalty, got %d", result.TotalPenalty)
	}
}

func TestManager_Clear(t *testing.T) {
	manager := NewManager()

	manager.Register(&MockConstraint{name: "test", typ: Type("test"), category: CategoryHard})
	manager.Clear()

	if len(manager.GetAll()) != 0 {
		t.Error("Expected 0 constraints after clear")
	}
}

func TestManager_Count(t *testing.T) {
	manager := NewManager()

	if manager.Count() != 0 {
		t.Error("Expected 0 count for empty manager")
	}

	manager.Register(&MockConstraint{name: "c1", typ: Type("c1"), category: CategoryHard})
	manager.Register(&MockConstraint{name: "c2", typ: Type("c2"), category: CategorySoft})

	if manager.Count() != 2 {
		t.Errorf("Expected 2 count, got %d", manager.Count())
	}
}

// MockConstraint 用于测试的模拟约束
type MockConstraint struct {
	name     string
	typ      Type
	category Category
	weight   int
	pass     bool
	penalty  int
}

func (m *MockConstraint) Name() string       { return m.name }
func (m *MockConstraint) Type() Type         { return m.typ }
func (m *MockConstraint) Category() Category { return m.category }
func (m *MockConstraint) Weight() int {
	if m.weight == 0 {
		return 100
	}
	return m.weight
}

func (m *MockConstraint) Evaluate(ctx *Context) (bool, int, []ViolationDetail) {
	if m.pass {
		return true, 0, nil
	}
	return false, m.penalty, []ViolationDetail{
		{ConstraintName: m.name, Message: "违反约束", Penalty: m.penalty},
	}
}

func (m *MockConstraint) EvaluateAssignment(ctx *Context, assignment *model.Assignment) (bool, int) {
	return m.pass, m.penalty
}
