// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"

	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// RegisterDefaultConstraints 注册默认约束到管理器
func RegisterDefaultConstraints(manager *constraint.Manager, config map[string]interface{}) {
	// 从配置中获取参数，使用默认值
	maxHoursPerDay := getConfigInt(config, "max_hours_per_day", 10)
	maxHoursPerWeek := getConfigInt(config, "max_hours_per_week", 44)
	maxHoursPerPeriod := getConfigInt(config, "max_hours_per_period", 0) // 0表示不限制
	maxShiftsPerMonth := getConfigInt(config, "max_shifts_per_month", 0) // 0表示不限制
	minRestBetweenShifts := getConfigInt(config, "min_rest_between_shifts", 10)
	maxConsecutiveDays := getConfigInt(config, "max_consecutive_days", 6)
	standardHoursPerWeek := getConfigInt(config, "standard_hours_per_week", 40)
	workloadBalanceWeight := getConfigInt(config, "workload_balance_weight", 60)
	preferenceWeight := getConfigInt(config, "preference_weight", 50)
	minimizeOvertimeWeight := getConfigInt(config, "minimize_overtime_weight", 70)
	tolerancePercent := getConfigFloat(config, "workload_tolerance_percent", 20.0)

	// 工时模式: "weekly"(按周) 或 "period"(按排班周期)
	hoursMode := getConfigString(config, "hours_mode", "weekly")

	// 注册硬约束
	manager.Register(NewMaxHoursPerDayConstraint(maxHoursPerDay))

	// 根据工时模式选择约束
	if hoursMode == "period" && maxHoursPerPeriod > 0 {
		// 按排班周期计算工时（适用于月度排班）
		manager.Register(NewMaxHoursPerPeriodConstraint(maxHoursPerPeriod))
	} else {
		// 按周计算工时（默认模式）
		manager.Register(NewMaxHoursPerWeekConstraint(maxHoursPerWeek))
	}

	manager.Register(NewMinRestBetweenShiftsConstraint(minRestBetweenShifts))
	manager.Register(NewMaxConsecutiveDaysConstraint(maxConsecutiveDays))
	manager.Register(NewMaxShiftsPerDayConstraint(1)) // 每天最多1个班次
	manager.Register(NewSkillRequiredConstraint())

	// 每月最大班次数约束（如果配置了）
	if maxShiftsPerMonth > 0 {
		// 获取每月单独设置的限制（可选）
		// 格式: { "2026-01": 20, "2026-02": 26, ... }
		monthlyLimits := getConfigMonthlyLimits(config, "monthly_max_shifts")
		
		manager.Register(NewMaxShiftsPerMonthConstraint(maxShiftsPerMonth, monthlyLimits))
		if len(monthlyLimits) > 0 {
			fmt.Printf("📅 已注册每月最大班次数约束: 默认%d班/月, 自定义月份: %v\n", maxShiftsPerMonth, monthlyLimits)
		} else {
			fmt.Printf("📅 已注册每月最大班次数约束: %d 班/月\n", maxShiftsPerMonth)
		}
	}

	// 注册软约束
	manager.Register(NewWorkloadBalanceConstraint(workloadBalanceWeight, tolerancePercent))
	manager.Register(NewEmployeePreferenceConstraint(preferenceWeight))
	manager.Register(NewMinimizeOvertimeConstraint(minimizeOvertimeWeight, standardHoursPerWeek))
}

// RegisterRestaurantConstraints 注册餐饮场景约束
func RegisterRestaurantConstraints(manager *constraint.Manager, config map[string]interface{}) {
	// 首先注册默认约束
	RegisterDefaultConstraints(manager, config)

	// 餐饮行业资质要求（健康证等）
	manager.Register(NewIndustryCertificationConstraint("restaurant"))

	// 高峰期覆盖
	peakHours := []string{"11:00-13:00", "17:00-20:00"}
	if ph, ok := config["peak_hours"].([]string); ok {
		peakHours = ph
	}
	minPeakStaff := getConfigInt(config, "min_peak_staff", 3)
	manager.Register(NewPeakHoursCoverageConstraint(90, peakHours, minPeakStaff))

	// 两头班支持
	maxSplitShifts := getConfigInt(config, "max_split_shifts_per_week", 2)
	allowSplit := true
	if allow, ok := config["allow_split_shift"].(bool); ok {
		allowSplit = allow
	}
	manager.Register(NewSplitShiftConstraint(60, maxSplitShifts, 3, allowSplit))
}

// RegisterFactoryConstraints 注册工厂场景约束
func RegisterFactoryConstraints(manager *constraint.Manager, config map[string]interface{}) {
	// 首先注册默认约束
	RegisterDefaultConstraints(manager, config)

	// 工厂特种作业资质要求
	manager.Register(NewIndustryCertificationConstraint("factory"))

	// 倒班模式
	pattern := getConfigString(config, "shift_rotation_pattern", "三班倒")
	rotationDays := getConfigInt(config, "rotation_days", 7)
	manager.Register(NewShiftRotationPatternConstraint(100, pattern, rotationDays))

	// 最大连续夜班
	maxNights := getConfigInt(config, "max_consecutive_nights", 4)
	manager.Register(NewMaxConsecutiveNightsConstraint(maxNights))
}

// getConfigString 从配置中获取字符串
func getConfigString(config map[string]interface{}, key string, defaultVal string) string {
	if config == nil {
		return defaultVal
	}
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultVal
}

// RegisterHousekeepingConstraints 注册家政场景约束
func RegisterHousekeepingConstraints(manager *constraint.Manager, config map[string]interface{}) {
	// 首先注册默认约束
	RegisterDefaultConstraints(manager, config)

	// 家政行业资质要求（无犯罪证明等）
	manager.Register(NewIndustryCertificationConstraint("housekeeping"))

	// 家政特有约束
	// 服务区域匹配（硬约束）
	manager.Register(NewServiceAreaMatchConstraint())

	// 路程时间缓冲（软约束）
	travelBuffer := getConfigInt(config, "travel_buffer_minutes", 30)
	manager.Register(NewTravelTimeBufferConstraint(travelBuffer))

	// 客户偏好（软约束）
	preferenceWeight := getConfigInt(config, "customer_preference_weight", 50)
	manager.Register(NewCustomerPreferenceConstraint(preferenceWeight))
}

// RegisterNursingConstraints 注册长护险场景约束
func RegisterNursingConstraints(manager *constraint.Manager, config map[string]interface{}) {
	// 首先注册默认约束
	RegisterDefaultConstraints(manager, config)

	// 长护险资质要求（护理员证、无犯罪证明等）
	manager.Register(NewIndustryCertificationConstraint("nursing"))

	// 长护险特有约束
	// 护理计划合规（硬约束）
	manager.Register(NewCarePlanComplianceConstraint())

	// 护理员连续性（软约束）
	continuityWeight := getConfigInt(config, "caregiver_continuity_weight", 85)
	manager.Register(NewCaregiverContinuityConstraint(continuityWeight))

	// 服务时间规律性（软约束）
	regularityWeight := getConfigInt(config, "service_regularity_weight", 60)
	manager.Register(NewServiceTimeRegularityConstraint(regularityWeight))

	// 每日最大服务患者数（硬约束）
	maxPatients := getConfigInt(config, "max_patients_per_day", 4)
	manager.Register(NewMaxPatientsPerDayConstraint(maxPatients))
}

// getConfigInt 从配置中获取整数
func getConfigInt(config map[string]interface{}, key string, defaultVal int) int {
	if config == nil {
		return defaultVal
	}
	if val, ok := config[key]; ok {
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

// getConfigFloat 从配置中获取浮点数
func getConfigFloat(config map[string]interface{}, key string, defaultVal float64) float64 {
	if config == nil {
		return defaultVal
	}
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		}
	}
	return defaultVal
}

// getConfigMonthlyLimits 从配置中获取每月限制的 map
// 格式: { "2026-01": 20, "2026-02": 26, ... }
func getConfigMonthlyLimits(config map[string]interface{}, key string) map[string]int {
	result := make(map[string]int)
	if config == nil {
		return result
	}
	
	val, ok := config[key]
	if !ok {
		return result
	}
	
	// 尝试转换为 map[string]interface{}
	if m, ok := val.(map[string]interface{}); ok {
		for month, limit := range m {
			switch v := limit.(type) {
			case int:
				result[month] = v
			case float64:
				result[month] = int(v)
			case int64:
				result[month] = int(v)
			}
		}
	}
	
	return result
}
