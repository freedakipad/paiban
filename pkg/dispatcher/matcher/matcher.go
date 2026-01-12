// Package matcher 提供技能和距离匹配功能
package matcher

import (
	"math"
	"sort"

	"github.com/paiban/paiban/pkg/model"
)

// MatchScore 匹配评分
type MatchScore struct {
	EmployeeID    string   `json:"employee_id"`
	EmployeeName  string   `json:"employee_name"`
	TotalScore    float64  `json:"total_score"`
	SkillScore    float64  `json:"skill_score"`
	DistanceScore float64  `json:"distance_score"`
	HistoryScore  float64  `json:"history_score"`
	MatchedSkills []string `json:"matched_skills"`
	Distance      float64  `json:"distance_km"`
}

// SkillMatcher 技能匹配器
type SkillMatcher struct {
	skillWeights map[string]float64 // 技能权重
}

// NewSkillMatcher 创建技能匹配器
func NewSkillMatcher() *SkillMatcher {
	return &SkillMatcher{
		skillWeights: map[string]float64{
			// 护理技能
			"基础护理": 1.0,
			"专业护理": 1.5,
			"康复训练": 1.8,
			"心理疏导": 1.3,
			"临终关怀": 2.0,
			// 家政技能
			"烹饪":   1.0,
			"保洁":   0.8,
			"婴儿护理": 1.5,
			"老人护理": 1.5,
			"月嫂服务": 1.8,
			// 默认权重
			"default": 1.0,
		},
	}
}

// MatchSkills 匹配技能
func (m *SkillMatcher) MatchSkills(requiredSkills []string, employee *model.Employee) (float64, []string) {
	if len(requiredSkills) == 0 {
		return 100, nil
	}

	empSkills := make(map[string]bool)
	for _, s := range employee.Skills {
		empSkills[s] = true
	}

	matchedSkills := make([]string, 0)
	totalWeight := 0.0
	matchedWeight := 0.0

	for _, req := range requiredSkills {
		weight := m.skillWeights[req]
		if weight == 0 {
			weight = m.skillWeights["default"]
		}
		totalWeight += weight

		if empSkills[req] {
			matchedSkills = append(matchedSkills, req)
			matchedWeight += weight
		}
	}

	if totalWeight == 0 {
		return 100, matchedSkills
	}

	score := (matchedWeight / totalWeight) * 100
	return score, matchedSkills
}

// DistanceMatcher 距离匹配器
type DistanceMatcher struct {
	maxDistanceKm float64
}

// NewDistanceMatcher 创建距离匹配器
func NewDistanceMatcher(maxDistance float64) *DistanceMatcher {
	return &DistanceMatcher{
		maxDistanceKm: maxDistance,
	}
}

// CalculateDistance 计算两点距离
func (d *DistanceMatcher) CalculateDistance(loc1, loc2 *model.Location) float64 {
	if loc1 == nil || loc2 == nil {
		return 0
	}
	return haversineDistance(loc1.Latitude, loc1.Longitude, loc2.Latitude, loc2.Longitude)
}

// ScoreDistance 距离评分（距离越近分数越高）
func (d *DistanceMatcher) ScoreDistance(distance float64) float64 {
	if distance <= 0 {
		return 100
	}
	if distance >= d.maxDistanceKm {
		return 0
	}
	return (1 - distance/d.maxDistanceKm) * 100
}

// haversineDistance Haversine公式计算距离
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371

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

// HistoryMatcher 历史服务匹配器
type HistoryMatcher struct{}

// NewHistoryMatcher 创建历史匹配器
func NewHistoryMatcher() *HistoryMatcher {
	return &HistoryMatcher{}
}

// ScoreHistory 根据服务历史评分
func (h *HistoryMatcher) ScoreHistory(employeeID string, history []model.CustomerEmployeeHistory) float64 {
	for _, hist := range history {
		if hist.EmployeeID.String() == employeeID {
			score := 0.0

			// 服务次数评分
			if hist.ServiceCount >= 10 {
				score += 30
			} else if hist.ServiceCount >= 5 {
				score += 20
			} else if hist.ServiceCount >= 2 {
				score += 10
			}

			// 评价评分
			if hist.AvgRating >= 4.8 {
				score += 30
			} else if hist.AvgRating >= 4.5 {
				score += 20
			} else if hist.AvgRating >= 4.0 {
				score += 10
			}

			// 主护理员加分
			if hist.IsPrimary {
				score += 20
			}

			return score
		}
	}

	return 0
}

// ComprehensiveMatcher 综合匹配器
type ComprehensiveMatcher struct {
	skillMatcher    *SkillMatcher
	distanceMatcher *DistanceMatcher
	historyMatcher  *HistoryMatcher

	// 权重配置
	skillWeight    float64
	distanceWeight float64
	historyWeight  float64
}

// NewComprehensiveMatcher 创建综合匹配器
func NewComprehensiveMatcher(maxDistance float64) *ComprehensiveMatcher {
	return &ComprehensiveMatcher{
		skillMatcher:    NewSkillMatcher(),
		distanceMatcher: NewDistanceMatcher(maxDistance),
		historyMatcher:  NewHistoryMatcher(),
		skillWeight:     0.4,
		distanceWeight:  0.35,
		historyWeight:   0.25,
	}
}

// SetWeights 设置权重
func (c *ComprehensiveMatcher) SetWeights(skill, distance, history float64) {
	total := skill + distance + history
	c.skillWeight = skill / total
	c.distanceWeight = distance / total
	c.historyWeight = history / total
}

// Match 综合匹配
func (c *ComprehensiveMatcher) Match(order *model.ServiceOrder, employees []*model.Employee, customer *model.Customer, history []model.CustomerEmployeeHistory) []MatchScore {
	scores := make([]MatchScore, 0, len(employees))

	for _, emp := range employees {
		score := c.scoreEmployee(order, emp, customer, history)
		scores = append(scores, score)
	}

	// 按总分排序（分数越高越好）
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].TotalScore > scores[j].TotalScore
	})

	return scores
}

// scoreEmployee 评估单个员工
func (c *ComprehensiveMatcher) scoreEmployee(order *model.ServiceOrder, emp *model.Employee, customer *model.Customer, history []model.CustomerEmployeeHistory) MatchScore {
	// 技能评分
	skillScore, matchedSkills := c.skillMatcher.MatchSkills(order.Skills, emp)

	// 距离评分
	var distance float64
	var distanceScore float64 = 100
	if order.Location != nil && emp.HomeLocation != nil {
		// 使用员工的家庭位置计算到工作地点的距离
		distance = emp.HomeLocation.Distance(*order.Location)
		distanceScore = c.distanceMatcher.ScoreDistance(distance)
	}

	// 历史评分
	historyScore := c.historyMatcher.ScoreHistory(emp.ID.String(), history)

	// 综合评分
	totalScore := skillScore*c.skillWeight + distanceScore*c.distanceWeight + historyScore*c.historyWeight

	return MatchScore{
		EmployeeID:    emp.ID.String(),
		EmployeeName:  emp.Name,
		TotalScore:    totalScore,
		SkillScore:    skillScore,
		DistanceScore: distanceScore,
		HistoryScore:  historyScore,
		MatchedSkills: matchedSkills,
		Distance:      distance,
	}
}

// FindBestMatch 找到最佳匹配
func (c *ComprehensiveMatcher) FindBestMatch(order *model.ServiceOrder, employees []*model.Employee, customer *model.Customer, history []model.CustomerEmployeeHistory) *MatchScore {
	scores := c.Match(order, employees, customer, history)
	if len(scores) == 0 {
		return nil
	}
	return &scores[0]
}
