// Package scenario 提供场景测试
package scenario

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
	"github.com/paiban/paiban/pkg/scheduler/constraint/builtin"
)

// TestRestaurantHealthCertification 餐饮健康证约束测试
func TestRestaurantHealthCertification(t *testing.T) {
	cm := constraint.NewManager()
	cm.Register(builtin.NewIndustryCertificationConstraint("restaurant"))

	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-15")

	// 创建有健康证的员工
	empWithCert := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "张三",
		Position:       "服务员",
		Certifications: []string{"健康证", "食品安全培训证"},
		Status:         "active",
	}

	// 创建没有健康证的员工
	empWithoutCert := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "李四",
		Position:       "服务员",
		Certifications: []string{}, // 没有任何证书
		Status:         "active",
	}

	ctx.SetEmployees([]*model.Employee{empWithCert, empWithoutCert})

	shift := createShift("早班", "M", "08:00", "16:00", 480, "morning")
	ctx.SetShifts([]*model.Shift{shift})

	// 测试1：有健康证的员工 - 应该通过
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(empWithCert.ID, shift.ID, "2024-01-15", "08:00", "16:00"),
	})

	result := cm.Evaluate(ctx)
	if !result.IsValid {
		t.Error("有健康证的员工应该通过验证")
	}
	t.Logf("有健康证员工: 验证通过=%v, 违反数=%d", result.IsValid, len(result.HardViolations))

	// 测试2：没有健康证的员工 - 应该失败
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(empWithoutCert.ID, shift.ID, "2024-01-15", "08:00", "16:00"),
	})

	result = cm.Evaluate(ctx)
	if result.IsValid {
		t.Error("没有健康证的员工应该被拒绝")
	}
	if len(result.HardViolations) == 0 {
		t.Error("应该有硬约束违反记录")
	}

	t.Logf("无健康证员工: 验证通过=%v, 违反数=%d", result.IsValid, len(result.HardViolations))
	for _, v := range result.HardViolations {
		t.Logf("  违反: %s", v.Message)
	}
}

// TestHousekeepingNoCriminalRecord 家政无犯罪证明约束测试
func TestHousekeepingNoCriminalRecord(t *testing.T) {
	cm := constraint.NewManager()
	cm.Register(builtin.NewIndustryCertificationConstraint("housekeeping"))

	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-15")

	// 创建有无犯罪证明的员工
	validEmployee := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "保姆王阿姨",
		Position:       "保姆",
		Certifications: []string{"无犯罪证明", "家政服务证"},
		Status:         "active",
	}

	// 创建没有无犯罪证明的员工
	invalidEmployee := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "新员工小李",
		Position:       "保姆",
		Certifications: []string{"家政服务证"}, // 有家政证但没有无犯罪证明
		Status:         "active",
	}

	ctx.SetEmployees([]*model.Employee{validEmployee, invalidEmployee})

	shift := createShift("全天", "D", "08:00", "17:00", 540, "regular")
	ctx.SetShifts([]*model.Shift{shift})

	// 测试有无犯罪证明的员工
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(validEmployee.ID, shift.ID, "2024-01-15", "08:00", "17:00"),
	})

	result := cm.Evaluate(ctx)
	if !result.IsValid {
		t.Error("有无犯罪证明的家政员工应该通过验证")
	}
	t.Logf("有证明员工: 验证通过=%v", result.IsValid)

	// 测试没有无犯罪证明的员工
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(invalidEmployee.ID, shift.ID, "2024-01-15", "08:00", "17:00"),
	})

	result = cm.Evaluate(ctx)
	if result.IsValid {
		t.Error("没有无犯罪证明的家政员工应该被拒绝")
	}

	t.Logf("无证明员工: 验证通过=%v, 违反数=%d", result.IsValid, len(result.HardViolations))
	for _, v := range result.HardViolations {
		t.Logf("  违反: %s", v.Message)
	}
}

// TestNursingCertification 长护险护理资质测试
func TestNursingCertification(t *testing.T) {
	cm := constraint.NewManager()
	cm.Register(builtin.NewIndustryCertificationConstraint("nursing"))

	orgID := uuid.New()
	ctx := constraint.NewContext(orgID, "2024-01-15", "2024-01-15")

	// 创建合格护理员
	qualifiedNurse := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "护理员张姐",
		Position:       "护理员",
		Certifications: []string{"无犯罪证明", "护理员证"},
		Status:         "active",
	}

	// 创建不合格护理员（缺少护理员证）
	unqualifiedNurse := &model.Employee{
		BaseModel:      model.BaseModel{ID: uuid.New()},
		Name:           "临时工小王",
		Position:       "护理员",
		Certifications: []string{"无犯罪证明"}, // 只有无犯罪证明，没有护理员证
		Status:         "active",
	}

	ctx.SetEmployees([]*model.Employee{qualifiedNurse, unqualifiedNurse})

	shift := createShift("护理班", "C", "08:00", "16:00", 480, "regular")
	ctx.SetShifts([]*model.Shift{shift})

	// 测试合格护理员
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(qualifiedNurse.ID, shift.ID, "2024-01-15", "08:00", "16:00"),
	})

	result := cm.Evaluate(ctx)
	if !result.IsValid {
		t.Error("有护理员证的员工应该通过验证")
	}
	t.Logf("合格护理员: 验证通过=%v", result.IsValid)

	// 测试不合格护理员
	ctx.SetAssignments([]*model.Assignment{
		createAssignment(unqualifiedNurse.ID, shift.ID, "2024-01-15", "08:00", "16:00"),
	})

	result = cm.Evaluate(ctx)
	if result.IsValid {
		t.Error("没有护理员证的员工应该被拒绝")
	}

	t.Logf("不合格护理员: 验证通过=%v", result.IsValid)
	for _, v := range result.HardViolations {
		t.Logf("  违反: %s", v.Message)
	}
}

// TestValidateCertificationsForScenario 资质验证辅助函数测试
func TestValidateCertificationsForScenario(t *testing.T) {
	tests := []struct {
		name       string
		scenario   string
		position   string
		certs      []string
		wantValid  bool
		wantMissing int
	}{
		{
			name:      "餐饮服务员有健康证",
			scenario:  "restaurant",
			position:  "服务员",
			certs:     []string{"健康证"},
			wantValid: true,
		},
		{
			name:       "餐饮服务员无健康证",
			scenario:   "restaurant",
			position:   "服务员",
			certs:      []string{},
			wantValid:  false,
			wantMissing: 1,
		},
		{
			name:      "餐饮厨师完整证书",
			scenario:  "restaurant",
			position:  "厨师",
			certs:     []string{"健康证", "食品安全培训证"},
			wantValid: true,
		},
		{
			name:       "餐饮厨师缺少食品安全证",
			scenario:   "restaurant",
			position:   "厨师",
			certs:      []string{"健康证"},
			wantValid:  false,
			wantMissing: 1,
		},
		{
			name:      "家政保姆完整证书",
			scenario:  "housekeeping",
			position:  "保姆",
			certs:     []string{"无犯罪证明", "家政服务证"},
			wantValid: true,
		},
		{
			name:       "家政保姆缺少无犯罪证明",
			scenario:   "housekeeping",
			position:   "保姆",
			certs:      []string{"家政服务证"},
			wantValid:  false,
			wantMissing: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, missing := builtin.ValidateCertificationsForScenario(tt.scenario, tt.position, tt.certs)

			if valid != tt.wantValid {
				t.Errorf("ValidateCertificationsForScenario() valid = %v, want %v", valid, tt.wantValid)
			}

			if !tt.wantValid && len(missing) != tt.wantMissing {
				t.Errorf("missing count = %d, want %d, missing: %v", len(missing), tt.wantMissing, missing)
			}

			t.Logf("场景=%s, 岗位=%s, 有效=%v, 缺失=%v", tt.scenario, tt.position, valid, missing)
		})
	}
}

// TestGetScenarioCertRequirements 获取场景证书要求测试
func TestGetScenarioCertRequirements(t *testing.T) {
	scenarios := []string{"restaurant", "housekeeping", "nursing", "factory"}

	for _, scenario := range scenarios {
		reqs := builtin.GetScenarioCertRequirements(scenario)
		t.Logf("\n=== %s 场景证书要求 ===", scenario)
		for position, certs := range reqs {
			t.Logf("  %s: %v", position, certs)
		}

		if len(reqs) == 0 {
			t.Errorf("场景 %s 应该有证书要求", scenario)
		}
	}
}

