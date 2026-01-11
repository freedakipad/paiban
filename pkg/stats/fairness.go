// Package stats 提供排班统计分析功能
package stats

import (
	"math"
	"sort"
	"time"
)

// EmployeeInfo 员工信息（用于统计分析）
type EmployeeInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FairnessMetrics 公平性指标
type FairnessMetrics struct {
	// 工时公平性
	WorkloadGini           float64            `json:"workload_gini"`            // 工时基尼系数 (0=完全公平, 1=完全不公平)
	WorkloadVariance       float64            `json:"workload_variance"`        // 工时方差
	WorkloadStdDev         float64            `json:"workload_std_dev"`         // 工时标准差
	AvgHoursPerEmployee    float64            `json:"avg_hours_per_employee"`   // 人均工时
	MaxHours               float64            `json:"max_hours"`                // 最大工时
	MinHours               float64            `json:"min_hours"`                // 最小工时
	HoursRange             float64            `json:"hours_range"`              // 工时极差
	
	// 班次类型公平性
	ShiftTypeDistribution  map[string]float64 `json:"shift_type_distribution"`  // 各班次类型分布
	NightShiftGini         float64            `json:"night_shift_gini"`         // 夜班分配基尼系数
	WeekendShiftGini       float64            `json:"weekend_shift_gini"`       // 周末班分配基尼系数
	
	// 员工级别统计
	EmployeeStats          []EmployeeStat     `json:"employee_stats"`           // 员工统计
	
	// 综合评分
	OverallFairnessScore   float64            `json:"overall_fairness_score"`   // 综合公平性评分 (0-100)
}

// EmployeeStat 员工统计
type EmployeeStat struct {
	EmployeeID    string  `json:"employee_id"`
	EmployeeName  string  `json:"employee_name"`
	TotalHours    float64 `json:"total_hours"`
	ShiftCount    int     `json:"shift_count"`
	NightShifts   int     `json:"night_shifts"`
	WeekendShifts int     `json:"weekend_shifts"`
	OvertimeHours float64 `json:"overtime_hours"`
	Deviation     float64 `json:"deviation"` // 与平均值的偏差百分比
}

// FairnessAnalyzer 公平性分析器
type FairnessAnalyzer struct {
	standardWeeklyHours float64 // 标准周工时
	nightShiftStart     int     // 夜班开始时间（小时）
	nightShiftEnd       int     // 夜班结束时间（小时）
}

// NewFairnessAnalyzer 创建公平性分析器
func NewFairnessAnalyzer() *FairnessAnalyzer {
	return &FairnessAnalyzer{
		standardWeeklyHours: 40.0,
		nightShiftStart:     22,
		nightShiftEnd:       6,
	}
}

// Analyze 分析排班公平性
func (f *FairnessAnalyzer) Analyze(assignments []*AssignmentInfo, employees []*EmployeeInfo) *FairnessMetrics {
	if len(assignments) == 0 || len(employees) == 0 {
		return &FairnessMetrics{
			ShiftTypeDistribution: make(map[string]float64),
			OverallFairnessScore:  100,
		}
	}

	// 构建员工ID映射
	employeeMap := make(map[string]*EmployeeInfo)
	for _, e := range employees {
		employeeMap[e.ID] = e
	}

	// 统计每个员工的数据
	employeeStats := f.calculateEmployeeStats(assignments, employeeMap)

	// 计算工时列表
	hours := make([]float64, len(employeeStats))
	nightShifts := make([]float64, len(employeeStats))
	weekendShifts := make([]float64, len(employeeStats))
	
	for i, stat := range employeeStats {
		hours[i] = stat.TotalHours
		nightShifts[i] = float64(stat.NightShifts)
		weekendShifts[i] = float64(stat.WeekendShifts)
	}

	// 计算基本统计量
	avgHours := f.calculateMean(hours)
	variance := f.calculateVariance(hours, avgHours)
	stdDev := math.Sqrt(variance)
	maxHours, minHours := f.calculateRange(hours)

	// 更新员工偏差
	for i := range employeeStats {
		if avgHours > 0 {
			employeeStats[i].Deviation = (employeeStats[i].TotalHours - avgHours) / avgHours * 100
		}
	}

	// 计算基尼系数
	workloadGini := f.calculateGini(hours)
	nightGini := f.calculateGini(nightShifts)
	weekendGini := f.calculateGini(weekendShifts)

	// 计算班次类型分布
	shiftTypeDist := f.calculateShiftTypeDistribution(assignments)

	// 计算综合评分
	overallScore := f.calculateOverallScore(workloadGini, nightGini, weekendGini, stdDev, avgHours)

	return &FairnessMetrics{
		WorkloadGini:          workloadGini,
		WorkloadVariance:      variance,
		WorkloadStdDev:        stdDev,
		AvgHoursPerEmployee:   avgHours,
		MaxHours:              maxHours,
		MinHours:              minHours,
		HoursRange:            maxHours - minHours,
		ShiftTypeDistribution: shiftTypeDist,
		NightShiftGini:        nightGini,
		WeekendShiftGini:      weekendGini,
		EmployeeStats:         employeeStats,
		OverallFairnessScore:  overallScore,
	}
}

// calculateEmployeeStats 计算员工统计数据
func (f *FairnessAnalyzer) calculateEmployeeStats(assignments []*AssignmentInfo, employeeMap map[string]*EmployeeInfo) []EmployeeStat {
	statMap := make(map[string]*EmployeeStat)

	for _, a := range assignments {
		stat, exists := statMap[a.EmployeeID]
		if !exists {
			name := a.EmployeeID
			if e, ok := employeeMap[a.EmployeeID]; ok {
				name = e.Name
			}
			stat = &EmployeeStat{
				EmployeeID:   a.EmployeeID,
				EmployeeName: name,
			}
			statMap[a.EmployeeID] = stat
		}

		// 计算工时
		hours := f.calculateShiftHours(a.StartTime, a.EndTime)
		stat.TotalHours += hours
		stat.ShiftCount++

		// 检查是否是夜班
		if f.isNightShift(a.StartTime, a.EndTime) {
			stat.NightShifts++
		}

		// 检查是否是周末
		if f.isWeekend(a.Date) {
			stat.WeekendShifts++
		}
	}

	// 转换为切片
	result := make([]EmployeeStat, 0, len(statMap))
	for _, stat := range statMap {
		result = append(result, *stat)
	}

	// 按工时排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalHours > result[j].TotalHours
	})

	return result
}

// calculateShiftHours 计算班次工时
func (f *FairnessAnalyzer) calculateShiftHours(start, end time.Time) float64 {
	duration := end.Sub(start)
	return duration.Hours()
}

// isNightShift 判断是否是夜班
func (f *FairnessAnalyzer) isNightShift(start, end time.Time) bool {
	startHour := start.Hour()
	endHour := end.Hour()
	
	// 夜班定义：开始时间在22点后或结束时间在6点前
	return startHour >= f.nightShiftStart || endHour <= f.nightShiftEnd
}

// isWeekend 判断是否是周末
func (f *FairnessAnalyzer) isWeekend(dateStr string) bool {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// calculateMean 计算平均值
func (f *FairnessAnalyzer) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateVariance 计算方差
func (f *FairnessAnalyzer) calculateVariance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}

// calculateRange 计算极值
func (f *FairnessAnalyzer) calculateRange(values []float64) (max, min float64) {
	if len(values) == 0 {
		return 0, 0
	}
	max, min = values[0], values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	return
}

// calculateGini 计算基尼系数
func (f *FairnessAnalyzer) calculateGini(values []float64) float64 {
	n := len(values)
	if n == 0 {
		return 0
	}

	// 排序
	sorted := make([]float64, n)
	copy(sorted, values)
	sort.Float64s(sorted)

	// 计算累积和
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	if sum == 0 {
		return 0
	}

	// 计算基尼系数
	cumulativeSum := 0.0
	gini := 0.0
	for i, v := range sorted {
		cumulativeSum += v
		gini += (2*float64(i+1) - float64(n) - 1) * v
	}

	gini = gini / (float64(n) * sum)
	return math.Max(0, math.Min(1, gini))
}

// calculateShiftTypeDistribution 计算班次类型分布
func (f *FairnessAnalyzer) calculateShiftTypeDistribution(assignments []*AssignmentInfo) map[string]float64 {
	typeCounts := make(map[string]int)
	total := len(assignments)

	for _, a := range assignments {
		shiftType := f.classifyShiftType(a.StartTime, a.EndTime)
		typeCounts[shiftType]++
	}

	distribution := make(map[string]float64)
	if total > 0 {
		for shiftType, count := range typeCounts {
			distribution[shiftType] = float64(count) / float64(total) * 100
		}
	}

	return distribution
}

// classifyShiftType 分类班次类型
func (f *FairnessAnalyzer) classifyShiftType(start, end time.Time) string {
	startHour := start.Hour()
	
	if startHour >= 6 && startHour < 14 {
		return "morning"
	} else if startHour >= 14 && startHour < 22 {
		return "afternoon"
	} else {
		return "night"
	}
}

// calculateOverallScore 计算综合公平性评分
func (f *FairnessAnalyzer) calculateOverallScore(workloadGini, nightGini, weekendGini, stdDev, avgHours float64) float64 {
	// 各项权重
	const (
		workloadWeight = 0.4
		nightWeight    = 0.25
		weekendWeight  = 0.25
		stdDevWeight   = 0.1
	)

	// 基尼系数转换为分数 (0=100分, 1=0分)
	workloadScore := (1 - workloadGini) * 100
	nightScore := (1 - nightGini) * 100
	weekendScore := (1 - weekendGini) * 100

	// 标准差评分 (变异系数越低分数越高)
	cvScore := 100.0
	if avgHours > 0 {
		cv := stdDev / avgHours
		cvScore = math.Max(0, 100-cv*200)
	}

	// 加权平均
	score := workloadWeight*workloadScore +
		nightWeight*nightScore +
		weekendWeight*weekendScore +
		stdDevWeight*cvScore

	return math.Max(0, math.Min(100, score))
}

// CompareSchedules 比较两个排班方案的公平性
func (f *FairnessAnalyzer) CompareSchedules(schedule1, schedule2 []*AssignmentInfo, employees []*EmployeeInfo) map[string]float64 {
	metrics1 := f.Analyze(schedule1, employees)
	metrics2 := f.Analyze(schedule2, employees)

	return map[string]float64{
		"workload_gini_diff":      metrics2.WorkloadGini - metrics1.WorkloadGini,
		"night_gini_diff":         metrics2.NightShiftGini - metrics1.NightShiftGini,
		"weekend_gini_diff":       metrics2.WeekendShiftGini - metrics1.WeekendShiftGini,
		"overall_score_diff":      metrics2.OverallFairnessScore - metrics1.OverallFairnessScore,
		"schedule1_overall_score": metrics1.OverallFairnessScore,
		"schedule2_overall_score": metrics2.OverallFairnessScore,
	}
}

