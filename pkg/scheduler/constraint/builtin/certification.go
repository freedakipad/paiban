// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"

	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// 预定义证书类型常量
const (
	// 餐饮行业
	CertHealthCard      = "健康证"         // 餐饮从业人员健康证
	CertFoodSafety      = "食品安全培训证"   // 食品安全知识培训
	CertFireSafety      = "消防安全证"      // 消防安全培训

	// 家政服务
	CertNoCriminalRecord = "无犯罪证明"     // 无犯罪记录证明
	CertHousekeeping     = "家政服务证"     // 家政服务员资格证
	CertMaternityNurse   = "月嫂证"        // 母婴护理员证书
	CertElderlyCare      = "养老护理证"     // 养老护理员证书

	// 长护险/医疗护理
	CertNurseAide        = "护理员证"       // 护理员资格证
	CertNurseLicense     = "护士执照"       // 护士执业证书
	CertRehabTech        = "康复技师证"     // 康复技师证书
	CertCPR              = "心肺复苏证"     // CPR急救证书

	// 工厂
	CertElectrician      = "电工证"        // 电工操作证
	CertForklift         = "叉车证"        // 叉车操作证
	CertWelder           = "焊工证"        // 焊工操作证
	CertSafetyOfficer    = "安全员证"      // 安全生产管理员证
)

// IndustryCertificationConstraint 行业资质约束（硬约束）
// 根据不同行业和岗位要求员工必须持有特定证书
type IndustryCertificationConstraint struct {
	*BaseConstraint
	scenario       string            // 场景: restaurant/housekeeping/nursing/factory
	certRequirements map[string][]string // 岗位 -> 所需证书列表
}

// NewIndustryCertificationConstraint 创建行业资质约束
func NewIndustryCertificationConstraint(scenario string) *IndustryCertificationConstraint {
	c := &IndustryCertificationConstraint{
		BaseConstraint: NewBaseConstraint(
			"行业资质要求",
			constraint.TypeCertificationLevel,
			constraint.CategoryHard,
			100, // 硬约束，必须满足
		),
		scenario:         scenario,
		certRequirements: make(map[string][]string),
	}

	// 根据场景预设证书要求
	c.loadScenarioCertRequirements(scenario)

	return c
}

// loadScenarioCertRequirements 加载场景预设证书要求
func (c *IndustryCertificationConstraint) loadScenarioCertRequirements(scenario string) {
	switch scenario {
	case "restaurant":
		// 餐饮行业：所有岗位都需要健康证
		c.certRequirements["服务员"] = []string{CertHealthCard}
		c.certRequirements["厨师"] = []string{CertHealthCard, CertFoodSafety}
		c.certRequirements["收银员"] = []string{CertHealthCard}
		c.certRequirements["厨师长"] = []string{CertHealthCard, CertFoodSafety, CertFireSafety}
		c.certRequirements["店长"] = []string{CertHealthCard, CertFireSafety}
		c.certRequirements["配菜员"] = []string{CertHealthCard}
		c.certRequirements["洗碗工"] = []string{CertHealthCard}
		// 默认：所有餐饮岗位需要健康证
		c.certRequirements["*"] = []string{CertHealthCard}

	case "housekeeping":
		// 家政服务：必须有无犯罪证明
		c.certRequirements["保姆"] = []string{CertNoCriminalRecord, CertHousekeeping}
		c.certRequirements["月嫂"] = []string{CertNoCriminalRecord, CertMaternityNurse}
		c.certRequirements["育儿嫂"] = []string{CertNoCriminalRecord, CertHousekeeping}
		c.certRequirements["保洁"] = []string{CertNoCriminalRecord}
		c.certRequirements["钟点工"] = []string{CertNoCriminalRecord}
		c.certRequirements["居家养老"] = []string{CertNoCriminalRecord, CertElderlyCare}
		// 默认：所有家政岗位需要无犯罪证明
		c.certRequirements["*"] = []string{CertNoCriminalRecord}

	case "nursing":
		// 长护险/护理：需要护理资质
		c.certRequirements["护理员"] = []string{CertNoCriminalRecord, CertNurseAide}
		c.certRequirements["护士"] = []string{CertNurseAide, CertNurseLicense, CertCPR}
		c.certRequirements["康复师"] = []string{CertNurseAide, CertRehabTech}
		c.certRequirements["生活护理"] = []string{CertNoCriminalRecord, CertNurseAide}
		// 默认：护理行业需要护理员证
		c.certRequirements["*"] = []string{CertNoCriminalRecord, CertNurseAide}

	case "factory":
		// 工厂：特种作业需要对应证书
		c.certRequirements["电工"] = []string{CertElectrician}
		c.certRequirements["焊工"] = []string{CertWelder}
		c.certRequirements["叉车工"] = []string{CertForklift}
		c.certRequirements["安全员"] = []string{CertSafetyOfficer}
		// 普通操作工不需要特殊证书
		c.certRequirements["操作工"] = []string{}
	}
}

// AddCertRequirement 添加自定义证书要求
func (c *IndustryCertificationConstraint) AddCertRequirement(position string, certs []string) {
	c.certRequirements[position] = certs
}

// Evaluate 评估整个排班
func (c *IndustryCertificationConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, a := range ctx.Assignments {
		emp := ctx.GetEmployee(a.EmployeeID)
		if emp == nil {
			continue
		}

		// 确定岗位
		position := a.Position
		if position == "" {
			position = emp.Position
		}

		// 获取该岗位所需证书
		requiredCerts := c.getRequiredCerts(position)

		// 检查员工是否持有所有必需证书
		for _, cert := range requiredCerts {
			if !emp.HasCertification(cert) {
				isValid = false
				penalty := c.Weight()
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           a.Date,
					Message: fmt.Sprintf(
						"[%s场景] 员工 %s 岗位 '%s' 缺少必需证书: %s",
						c.getScenarioName(), emp.Name, position, cert,
					),
					Severity: "error",
					Penalty:  penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// getRequiredCerts 获取岗位所需证书
func (c *IndustryCertificationConstraint) getRequiredCerts(position string) []string {
	// 先查找具体岗位
	if certs, ok := c.certRequirements[position]; ok {
		return certs
	}
	// 再查找默认要求
	if certs, ok := c.certRequirements["*"]; ok {
		return certs
	}
	return nil
}

// getScenarioName 获取场景中文名
func (c *IndustryCertificationConstraint) getScenarioName() string {
	switch c.scenario {
	case "restaurant":
		return "餐饮"
	case "housekeeping":
		return "家政"
	case "nursing":
		return "护理"
	case "factory":
		return "工厂"
	default:
		return c.scenario
	}
}

// EvaluateAssignment 评估单个分配
func (c *IndustryCertificationConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	emp := ctx.GetEmployee(a.EmployeeID)
	if emp == nil {
		return false, c.Weight()
	}

	position := a.Position
	if position == "" {
		position = emp.Position
	}

	requiredCerts := c.getRequiredCerts(position)
	for _, cert := range requiredCerts {
		if !emp.HasCertification(cert) {
			return false, c.Weight()
		}
	}

	return true, 0
}

// ValidateCertificationsForScenario 验证员工是否满足场景资质要求
// 可用于入职检查或排班前验证
func ValidateCertificationsForScenario(scenario, position string, certifications []string) (bool, []string) {
	c := NewIndustryCertificationConstraint(scenario)
	requiredCerts := c.getRequiredCerts(position)

	// 构建持有证书集合
	certSet := make(map[string]bool)
	for _, cert := range certifications {
		certSet[cert] = true
	}

	// 检查缺失证书
	var missing []string
	for _, required := range requiredCerts {
		if !certSet[required] {
			missing = append(missing, required)
		}
	}

	return len(missing) == 0, missing
}

// GetScenarioCertRequirements 获取场景的证书要求（用于API查询）
func GetScenarioCertRequirements(scenario string) map[string][]string {
	c := NewIndustryCertificationConstraint(scenario)
	return c.certRequirements
}

