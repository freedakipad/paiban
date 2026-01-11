// Package careplan 提供长护险护理计划管理
package careplan

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

// PlanManager 护理计划管理器
type PlanManager struct {
	// 护理等级对应的周服务时长（小时）
	levelHours map[int]int
	// 护理等级对应的服务项目
	levelItems map[int][]string
}

// NewPlanManager 创建护理计划管理器
func NewPlanManager() *PlanManager {
	return &PlanManager{
		levelHours: map[int]int{
			1: 3,  // 一级：每周3小时
			2: 5,  // 二级：每周5小时
			3: 7,  // 三级：每周7小时
			4: 10, // 四级：每周10小时
			5: 15, // 五级：每周15小时
			6: 20, // 六级：每周20小时
		},
		levelItems: map[int][]string{
			1: {"基础生活照料", "健康监测"},
			2: {"基础生活照料", "健康监测", "饮食护理"},
			3: {"基础生活照料", "健康监测", "饮食护理", "排泄护理"},
			4: {"基础生活照料", "健康监测", "饮食护理", "排泄护理", "清洁护理"},
			5: {"基础生活照料", "健康监测", "饮食护理", "排泄护理", "清洁护理", "康复训练"},
			6: {"基础生活照料", "健康监测", "饮食护理", "排泄护理", "清洁护理", "康复训练", "临终关怀"},
		},
	}
}

// CreatePlan 创建护理计划
func (pm *PlanManager) CreatePlan(customerID uuid.UUID, level int, startDate string) (*model.CarePlan, error) {
	if level < 1 || level > 6 {
		return nil, fmt.Errorf("护理等级必须在1-6之间")
	}

	weeklyHours, ok := pm.levelHours[level]
	if !ok {
		weeklyHours = 5
	}

	itemNames := pm.levelItems[level]
	items := make([]model.CareItem, len(itemNames))
	for i, name := range itemNames {
		items[i] = model.CareItem{
			Code:      fmt.Sprintf("CARE_%d_%d", level, i+1),
			Name:      name,
			Frequency: 2,                                     // 每周2次
			Duration:  weeklyHours * 60 / 2 / len(itemNames), // 平均分配时长
		}
	}

	plan := &model.CarePlan{
		CustomerID:   customerID,
		PlanNo:       generatePlanNo(),
		Level:        level,
		StartDate:    startDate,
		WeeklyHours:  weeklyHours,
		ServiceItems: items,
		Frequency:    calculateFrequency(weeklyHours),
		Status:       "active",
	}

	return plan, nil
}

// GenerateServiceOrders 根据护理计划生成服务订单
func (pm *PlanManager) GenerateServiceOrders(plan *model.CarePlan, customer *model.Customer, startDate, endDate string) ([]*model.ServiceOrder, error) {
	if plan == nil || plan.Status != "active" {
		return nil, fmt.Errorf("护理计划无效或已过期")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("开始日期格式错误: %v", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("结束日期格式错误: %v", err)
	}

	var orders []*model.ServiceOrder

	// 计算每周服务次数
	sessionsPerWeek := calculateSessionsPerWeek(plan.WeeklyHours)
	sessionDuration := plan.WeeklyHours * 60 / sessionsPerWeek // 每次服务时长（分钟）

	// 服务日安排（均匀分布在一周内）
	serviceDays := getServiceDays(sessionsPerWeek)

	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		weekday := int(current.Weekday())

		// 检查是否是服务日
		isServiceDay := false
		for _, day := range serviceDays {
			if day == weekday {
				isServiceDay = true
				break
			}
		}

		if !isServiceDay {
			continue
		}

		// 创建订单
		order := &model.ServiceOrder{
			CustomerID:  plan.CustomerID,
			OrderNo:     generateOrderNo(current),
			ServiceType: "nursing",
			ServiceDate: current.Format("2006-01-02"),
			StartTime:   "09:00",
			EndTime:     calculateEndTime("09:00", sessionDuration),
			Duration:    sessionDuration,
			Status:      "pending",
			Priority:    3,
		}

		if customer != nil {
			order.Address = customer.Address
			order.Location = customer.Location
		}

		// 设置技能要求
		order.Skills = pm.getPlanSkillRequirements(plan.Level)

		orders = append(orders, order)
	}

	return orders, nil
}

// ValidatePlan 验证护理计划
func (pm *PlanManager) ValidatePlan(plan *model.CarePlan) []string {
	var errors []string

	if plan.Level < 1 || plan.Level > 6 {
		errors = append(errors, "护理等级无效")
	}

	if plan.WeeklyHours <= 0 {
		errors = append(errors, "周服务时长必须大于0")
	}

	if plan.StartDate == "" {
		errors = append(errors, "开始日期不能为空")
	}

	if len(plan.ServiceItems) == 0 {
		errors = append(errors, "服务项目不能为空")
	}

	return errors
}

// GetRecommendedCarers 获取推荐护理员
func (pm *PlanManager) GetRecommendedCarers(plan *model.CarePlan, carers []*model.Employee) []*CarerRecommendation {
	if plan == nil || len(carers) == 0 {
		return nil
	}

	// 获取护理等级所需技能
	requiredSkills := pm.getPlanSkillRequirements(plan.Level)

	recommendations := make([]*CarerRecommendation, 0)

	for _, carer := range carers {
		score := 0.0
		matchedSkills := []string{}

		// 技能匹配评分
		carerSkills := make(map[string]bool)
		for _, s := range carer.Skills {
			carerSkills[s] = true
		}

		for _, req := range requiredSkills {
			if carerSkills[req] {
				matchedSkills = append(matchedSkills, req)
				score += 20
			}
		}

		// 证书检查
		hasCert := false
		for _, cert := range carer.Certifications {
			if cert == "护理员证" {
				hasCert = true
				score += 30
				break
			}
		}

		if !hasCert {
			continue // 无资质不推荐
		}

		// 状态检查
		if carer.Status != "active" {
			continue
		}

		recommendations = append(recommendations, &CarerRecommendation{
			Carer:         carer,
			Score:         score,
			MatchedSkills: matchedSkills,
			Suitable:      len(matchedSkills) >= len(requiredSkills)/2,
		})
	}

	return recommendations
}

// getPlanSkillRequirements 获取护理计划所需技能
func (pm *PlanManager) getPlanSkillRequirements(level int) []string {
	skills := []string{"护理员证", "健康证"}

	switch level {
	case 5, 6:
		skills = append(skills, "康复训练", "专业护理")
	case 3, 4:
		skills = append(skills, "基础护理")
	}

	return skills
}

// CarerRecommendation 护理员推荐
type CarerRecommendation struct {
	Carer         *model.Employee `json:"carer"`
	Score         float64         `json:"score"`
	MatchedSkills []string        `json:"matched_skills"`
	Suitable      bool            `json:"suitable"`
}

// 辅助函数

func generatePlanNo() string {
	return fmt.Sprintf("CP%s%04d", time.Now().Format("20060102"), time.Now().UnixNano()%10000)
}

func generateOrderNo(date time.Time) string {
	return fmt.Sprintf("SO%s%06d", date.Format("20060102"), time.Now().UnixNano()%1000000)
}

func calculateFrequency(weeklyHours int) string {
	if weeklyHours <= 3 {
		return "weekly"
	} else if weeklyHours <= 7 {
		return "twice_weekly"
	} else if weeklyHours <= 14 {
		return "three_times_weekly"
	}
	return "daily"
}

func calculateSessionsPerWeek(weeklyHours int) int {
	if weeklyHours <= 3 {
		return 1
	} else if weeklyHours <= 7 {
		return 2
	} else if weeklyHours <= 14 {
		return 3
	}
	return 5
}

func getServiceDays(sessionsPerWeek int) []int {
	switch sessionsPerWeek {
	case 1:
		return []int{3} // 周三
	case 2:
		return []int{2, 5} // 周二、周五
	case 3:
		return []int{1, 3, 5} // 周一、周三、周五
	default:
		return []int{1, 2, 3, 4, 5} // 工作日
	}
}

func calculateEndTime(start string, durationMinutes int) string {
	t, _ := time.Parse("15:04", start)
	t = t.Add(time.Duration(durationMinutes) * time.Minute)
	return t.Format("15:04")
}
