// Package constraint 提供派出服务约束
package constraint

import (
	"math"
	"time"

	"github.com/paiban/paiban/pkg/model"
)

// DispatchConstraint 派出服务约束接口
type DispatchConstraint interface {
	Name() string
	Type() string // hard/soft
	Weight() float64
	Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string)
}

// DispatchContext 派单上下文
type DispatchContext struct {
	Customer        *model.Customer
	TodayOrders     []*model.ServiceOrder   // 今日所有订单
	EmployeeOrders  []*model.ServiceOrder   // 员工今日已分配订单
	ServiceHistory  []model.CustomerEmployeeHistory // 客户服务历史
	EmployeeLocation *model.Location         // 员工当前位置
}

// BaseDispatchConstraint 基础派出约束
type BaseDispatchConstraint struct {
	name   string
	ctype  string
	weight float64
}

func (b *BaseDispatchConstraint) Name() string   { return b.name }
func (b *BaseDispatchConstraint) Type() string   { return b.ctype }
func (b *BaseDispatchConstraint) Weight() float64 { return b.weight }

// =========================================
// 1. ServiceAreaMatchConstraint 服务区域匹配
// =========================================
type ServiceAreaMatchConstraint struct {
	BaseDispatchConstraint
	MaxDistanceKm float64 // 最大服务距离
}

func NewServiceAreaMatchConstraint(maxDistance float64) *ServiceAreaMatchConstraint {
	return &ServiceAreaMatchConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "ServiceAreaMatch",
			ctype:  "hard",
			weight: 1000,
		},
		MaxDistanceKm: maxDistance,
	}
}

func (c *ServiceAreaMatchConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	if order.Location == nil || ctx.EmployeeLocation == nil {
		// 无位置信息，跳过检查
		return true, 0, ""
	}

	distance := calculateDistance(
		order.Location.Latitude, order.Location.Longitude,
		ctx.EmployeeLocation.Latitude, ctx.EmployeeLocation.Longitude,
	)

	if distance > c.MaxDistanceKm {
		return false, c.weight, "服务距离超出范围"
	}

	// 软惩罚：距离越远惩罚越高
	penalty := distance / c.MaxDistanceKm * 10
	return true, penalty, ""
}

// =========================================
// 2. TravelTimeBufferConstraint 路程时间缓冲
// =========================================
type TravelTimeBufferConstraint struct {
	BaseDispatchConstraint
	MinBufferMinutes int // 订单间最小缓冲时间
}

func NewTravelTimeBufferConstraint(minBuffer int) *TravelTimeBufferConstraint {
	return &TravelTimeBufferConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "TravelTimeBuffer",
			ctype:  "hard",
			weight: 500,
		},
		MinBufferMinutes: minBuffer,
	}
}

func (c *TravelTimeBufferConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	if len(ctx.EmployeeOrders) == 0 {
		return true, 0, ""
	}

	orderStart, _ := time.Parse("15:04", order.StartTime)
	orderEnd, _ := time.Parse("15:04", order.EndTime)

	for _, existingOrder := range ctx.EmployeeOrders {
		existStart, _ := time.Parse("15:04", existingOrder.StartTime)
		existEnd, _ := time.Parse("15:04", existingOrder.EndTime)

		// 检查时间重叠
		if !(orderEnd.Before(existStart) || orderStart.After(existEnd)) {
			return false, c.weight, "与现有订单时间冲突"
		}

		// 检查缓冲时间
		buffer := 0
		if orderStart.After(existEnd) {
			buffer = int(orderStart.Sub(existEnd).Minutes())
		} else if existStart.After(orderEnd) {
			buffer = int(existStart.Sub(orderEnd).Minutes())
		}

		if buffer > 0 && buffer < c.MinBufferMinutes {
			return false, c.weight * 0.5, "订单间缓冲时间不足"
		}
	}

	return true, 0, ""
}

// =========================================
// 3. MaxOrdersPerDayConstraint 每日最大订单数
// =========================================
type MaxOrdersPerDayConstraint struct {
	BaseDispatchConstraint
	MaxOrders int
}

func NewMaxOrdersPerDayConstraint(maxOrders int) *MaxOrdersPerDayConstraint {
	return &MaxOrdersPerDayConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "MaxOrdersPerDay",
			ctype:  "hard",
			weight: 300,
		},
		MaxOrders: maxOrders,
	}
}

func (c *MaxOrdersPerDayConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	currentCount := len(ctx.EmployeeOrders)
	
	if currentCount >= c.MaxOrders {
		return false, c.weight, "员工今日订单数已满"
	}

	// 软惩罚：订单越多惩罚越高
	penalty := float64(currentCount) / float64(c.MaxOrders) * 5
	return true, penalty, ""
}

// =========================================
// 4. CustomerPreferenceConstraint 客户偏好
// =========================================
type CustomerPreferenceConstraint struct {
	BaseDispatchConstraint
}

func NewCustomerPreferenceConstraint() *CustomerPreferenceConstraint {
	return &CustomerPreferenceConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "CustomerPreference",
			ctype:  "soft",
			weight: 50,
		},
	}
}

func (c *CustomerPreferenceConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	if ctx.Customer == nil {
		return true, 0, ""
	}

	penalty := 0.0

	// 检查是否在黑名单中
	for _, blockedID := range ctx.Customer.BlockedEmpIDs {
		if blockedID == employee.ID {
			return false, 1000, "员工在客户黑名单中"
		}
	}

	// 检查是否是偏好员工
	isPreferred := false
	for _, prefID := range ctx.Customer.PreferredEmpIDs {
		if prefID == employee.ID {
			isPreferred = true
			penalty -= 20 // 奖励
			break
		}
	}

	// 检查客户偏好
	if ctx.Customer.Preferences != nil {
		prefs := ctx.Customer.Preferences
		
		// 要求同一服务者
		if prefs.RequireSameWorker && !isPreferred && len(ctx.ServiceHistory) > 0 {
			penalty += 30
		}
	}

	return true, penalty, ""
}

// =========================================
// 5. CertificationLevelConstraint 资质等级
// =========================================
type CertificationLevelConstraint struct {
	BaseDispatchConstraint
	ServiceCertRequirements map[string][]string // 服务类型 -> 所需证书
}

func NewCertificationLevelConstraint() *CertificationLevelConstraint {
	return &CertificationLevelConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "CertificationLevel",
			ctype:  "hard",
			weight: 800,
		},
		ServiceCertRequirements: map[string][]string{
			"nursing":     {"护理员证", "健康证"},
			"elder_care":  {"护理员证", "健康证"},
			"baby_care":   {"月嫂证", "健康证"},
			"cleaning":    {"无犯罪证明"},
			"cooking":     {"健康证"},
			"recovery":    {"康复技师证", "健康证"},
		},
	}
}

func (c *CertificationLevelConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	requiredCerts, exists := c.ServiceCertRequirements[order.ServiceType]
	if !exists {
		return true, 0, ""
	}

	empCerts := make(map[string]bool)
	for _, cert := range employee.Certifications {
		empCerts[cert] = true
	}

	for _, reqCert := range requiredCerts {
		if !empCerts[reqCert] {
			return false, c.weight, "缺少必需证书: " + reqCert
		}
	}

	return true, 0, ""
}

// =========================================
// 6. CaregiverContinuityConstraint 护理员连续性
// =========================================
type CaregiverContinuityConstraint struct {
	BaseDispatchConstraint
	PreferHistoryDays int // 优先考虑最近N天有服务历史的员工
}

func NewCaregiverContinuityConstraint() *CaregiverContinuityConstraint {
	return &CaregiverContinuityConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "CaregiverContinuity",
			ctype:  "soft",
			weight: 40,
		},
		PreferHistoryDays: 30,
	}
}

func (c *CaregiverContinuityConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	if len(ctx.ServiceHistory) == 0 {
		return true, 0, ""
	}

	// 查找该员工的服务历史
	for _, history := range ctx.ServiceHistory {
		if history.EmployeeID == employee.ID {
			// 计算奖励
			bonus := 0.0

			// 服务次数奖励
			if history.ServiceCount > 5 {
				bonus -= 15
			} else if history.ServiceCount > 2 {
				bonus -= 8
			}

			// 评分奖励
			if history.AvgRating >= 4.5 {
				bonus -= 10
			} else if history.AvgRating >= 4.0 {
				bonus -= 5
			}

			// 主护理员奖励
			if history.IsPrimary {
				bonus -= 20
			}

			return true, bonus, ""
		}
	}

	// 没有服务历史，轻微惩罚
	return true, 5, ""
}

// =========================================
// 7. SkillMatchConstraint 技能匹配
// =========================================
type SkillMatchConstraint struct {
	BaseDispatchConstraint
}

func NewSkillMatchConstraint() *SkillMatchConstraint {
	return &SkillMatchConstraint{
		BaseDispatchConstraint: BaseDispatchConstraint{
			name:   "SkillMatch",
			ctype:  "hard",
			weight: 600,
		},
	}
}

func (c *SkillMatchConstraint) Evaluate(order *model.ServiceOrder, employee *model.Employee, ctx *DispatchContext) (bool, float64, string) {
	if len(order.Skills) == 0 {
		return true, 0, ""
	}

	empSkills := make(map[string]bool)
	for _, skill := range employee.Skills {
		empSkills[skill] = true
	}

	for _, reqSkill := range order.Skills {
		if !empSkills[reqSkill] {
			return false, c.weight, "缺少必需技能: " + reqSkill
		}
	}

	return true, 0, ""
}

// =========================================
// 辅助函数
// =========================================

// calculateDistance 计算两点间距离（Haversine公式）
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // km

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// DefaultDispatchConstraints 返回默认派出约束集合
func DefaultDispatchConstraints() []DispatchConstraint {
	return []DispatchConstraint{
		NewServiceAreaMatchConstraint(20),        // 最大20km
		NewTravelTimeBufferConstraint(30),        // 最小30分钟缓冲
		NewMaxOrdersPerDayConstraint(8),          // 每日最多8单
		NewCustomerPreferenceConstraint(),        // 客户偏好
		NewCertificationLevelConstraint(),        // 资质检查
		NewCaregiverContinuityConstraint(),       // 连续性偏好
		NewSkillMatchConstraint(),                // 技能匹配
	}
}

