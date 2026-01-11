// Package validator 提供排班验证功能
package validator

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

// ConflictType 冲突类型
type ConflictType string

const (
	ConflictOverlap     ConflictType = "overlap"       // 时间重叠
	ConflictRestTime    ConflictType = "rest_time"     // 休息时间不足
	ConflictMaxHours    ConflictType = "max_hours"     // 超过最大工时
	ConflictConsecutive ConflictType = "consecutive"   // 连续天数过多
	ConflictSkill       ConflictType = "skill"         // 技能不匹配
	ConflictAvailability ConflictType = "availability" // 不可用
)

// Conflict 冲突信息
type Conflict struct {
	Type        ConflictType  `json:"type"`
	Severity    string        `json:"severity"` // error/warning
	EmployeeID  uuid.UUID     `json:"employee_id"`
	Date        string        `json:"date"`
	Message     string        `json:"message"`
	Assignments []uuid.UUID   `json:"assignments,omitempty"` // 相关的排班ID
}

// ConflictDetector 冲突检测器
type ConflictDetector struct {
	config *DetectorConfig
}

// DetectorConfig 检测器配置
type DetectorConfig struct {
	MinRestHours       int  // 最小休息时间（小时）
	MaxHoursPerDay     int  // 每日最大工时
	MaxHoursPerWeek    int  // 每周最大工时
	MaxConsecutiveDays int  // 最大连续工作天数
	CheckSkills        bool // 是否检查技能
	CheckAvailability  bool // 是否检查可用性
}

// DefaultDetectorConfig 返回默认配置
func DefaultDetectorConfig() *DetectorConfig {
	return &DetectorConfig{
		MinRestHours:       10,
		MaxHoursPerDay:     10,
		MaxHoursPerWeek:    44,
		MaxConsecutiveDays: 6,
		CheckSkills:        true,
		CheckAvailability:  true,
	}
}

// NewConflictDetector 创建冲突检测器
func NewConflictDetector(config *DetectorConfig) *ConflictDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}
	return &ConflictDetector{config: config}
}

// DetectAll 检测所有冲突
func (d *ConflictDetector) DetectAll(assignments []*model.Assignment, employees map[uuid.UUID]*model.Employee) []Conflict {
	var conflicts []Conflict

	// 按员工分组
	byEmployee := groupByEmployee(assignments)

	for empID, empAssignments := range byEmployee {
		emp := employees[empID]
		if emp == nil {
			continue
		}

		// 检测各类冲突
		conflicts = append(conflicts, d.detectOverlaps(emp, empAssignments)...)
		conflicts = append(conflicts, d.detectRestTimeViolations(emp, empAssignments)...)
		conflicts = append(conflicts, d.detectMaxHoursViolations(emp, empAssignments)...)
		conflicts = append(conflicts, d.detectConsecutiveDaysViolations(emp, empAssignments)...)
	}

	return conflicts
}

// DetectForAssignment 检测单个分配的冲突
func (d *ConflictDetector) DetectForAssignment(
	newAssignment *model.Assignment,
	existingAssignments []*model.Assignment,
	employee *model.Employee,
) []Conflict {
	var conflicts []Conflict

	// 检测时间重叠
	for _, existing := range existingAssignments {
		if existing.EmployeeID != newAssignment.EmployeeID {
			continue
		}
		if existing.ID == newAssignment.ID {
			continue
		}

		if d.isOverlapping(newAssignment, existing) {
			conflicts = append(conflicts, Conflict{
				Type:        ConflictOverlap,
				Severity:    "error",
				EmployeeID:  newAssignment.EmployeeID,
				Date:        newAssignment.Date,
				Message:     fmt.Sprintf("与现有排班时间重叠"),
				Assignments: []uuid.UUID{newAssignment.ID, existing.ID},
			})
		}

		// 检测休息时间
		restHours := d.calculateRestHours(newAssignment, existing)
		if restHours >= 0 && restHours < float64(d.config.MinRestHours) {
			conflicts = append(conflicts, Conflict{
				Type:        ConflictRestTime,
				Severity:    "error",
				EmployeeID:  newAssignment.EmployeeID,
				Date:        newAssignment.Date,
				Message:     fmt.Sprintf("休息时间仅 %.1f 小时，少于要求的 %d 小时", restHours, d.config.MinRestHours),
				Assignments: []uuid.UUID{newAssignment.ID, existing.ID},
			})
		}
	}

	// 检测每日工时
	dailyHours := d.calculateDailyHours(newAssignment, existingAssignments)
	if dailyHours > float64(d.config.MaxHoursPerDay) {
		conflicts = append(conflicts, Conflict{
			Type:       ConflictMaxHours,
			Severity:   "error",
			EmployeeID: newAssignment.EmployeeID,
			Date:       newAssignment.Date,
			Message:    fmt.Sprintf("当日工时 %.1f 小时，超过限制 %d 小时", dailyHours, d.config.MaxHoursPerDay),
		})
	}

	return conflicts
}

// detectOverlaps 检测时间重叠
func (d *ConflictDetector) detectOverlaps(emp *model.Employee, assignments []*model.Assignment) []Conflict {
	var conflicts []Conflict

	// 按时间排序
	sorted := make([]*model.Assignment, len(assignments))
	copy(sorted, assignments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].StartTime.Before(sorted[j].StartTime)
	})

	// 检测相邻排班的重叠
	for i := 0; i < len(sorted)-1; i++ {
		current := sorted[i]
		next := sorted[i+1]

		if d.isOverlapping(current, next) {
			conflicts = append(conflicts, Conflict{
				Type:        ConflictOverlap,
				Severity:    "error",
				EmployeeID:  emp.ID,
				Date:        current.Date,
				Message:     fmt.Sprintf("员工 %s 在 %s 存在时间重叠的排班", emp.Name, current.Date),
				Assignments: []uuid.UUID{current.ID, next.ID},
			})
		}
	}

	return conflicts
}

// detectRestTimeViolations 检测休息时间不足
func (d *ConflictDetector) detectRestTimeViolations(emp *model.Employee, assignments []*model.Assignment) []Conflict {
	var conflicts []Conflict

	if len(assignments) < 2 {
		return conflicts
	}

	// 按结束时间排序
	sorted := make([]*model.Assignment, len(assignments))
	copy(sorted, assignments)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EndTime.Before(sorted[j].EndTime)
	})

	// 检查相邻班次间隔
	for i := 0; i < len(sorted)-1; i++ {
		current := sorted[i]
		next := sorted[i+1]

		restHours := next.StartTime.Sub(current.EndTime).Hours()
		if restHours >= 0 && restHours < float64(d.config.MinRestHours) {
			conflicts = append(conflicts, Conflict{
				Type:        ConflictRestTime,
				Severity:    "error",
				EmployeeID:  emp.ID,
				Date:        next.Date,
				Message:     fmt.Sprintf("员工 %s 班次间休息仅 %.1f 小时", emp.Name, restHours),
				Assignments: []uuid.UUID{current.ID, next.ID},
			})
		}
	}

	return conflicts
}

// detectMaxHoursViolations 检测工时超限
func (d *ConflictDetector) detectMaxHoursViolations(emp *model.Employee, assignments []*model.Assignment) []Conflict {
	var conflicts []Conflict

	// 按日期统计工时
	dailyHours := make(map[string]float64)
	var weeklyTotal float64

	for _, a := range assignments {
		hours := a.WorkingHours()
		dailyHours[a.Date] += hours
		weeklyTotal += hours
	}

	// 检查每日工时
	for date, hours := range dailyHours {
		if hours > float64(d.config.MaxHoursPerDay) {
			conflicts = append(conflicts, Conflict{
				Type:       ConflictMaxHours,
				Severity:   "error",
				EmployeeID: emp.ID,
				Date:       date,
				Message:    fmt.Sprintf("员工 %s 在 %s 工作 %.1f 小时，超过限制 %d 小时", emp.Name, date, hours, d.config.MaxHoursPerDay),
			})
		}
	}

	// 检查每周工时
	if weeklyTotal > float64(d.config.MaxHoursPerWeek) {
		conflicts = append(conflicts, Conflict{
			Type:       ConflictMaxHours,
			Severity:   "error",
			EmployeeID: emp.ID,
			Message:    fmt.Sprintf("员工 %s 周工作 %.1f 小时，超过限制 %d 小时", emp.Name, weeklyTotal, d.config.MaxHoursPerWeek),
		})
	}

	return conflicts
}

// detectConsecutiveDaysViolations 检测连续工作天数
func (d *ConflictDetector) detectConsecutiveDaysViolations(emp *model.Employee, assignments []*model.Assignment) []Conflict {
	var conflicts []Conflict

	if len(assignments) == 0 {
		return conflicts
	}

	// 获取工作日期
	workDates := make(map[string]bool)
	for _, a := range assignments {
		workDates[a.Date] = true
	}

	// 转换为排序列表
	dates := make([]string, 0, len(workDates))
	for d := range workDates {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	// 检测连续天数
	consecutive := 1
	maxConsecutive := 1
	startDate := dates[0]

	for i := 1; i < len(dates); i++ {
		if isConsecutiveDateStr(dates[i-1], dates[i]) {
			consecutive++
			if consecutive > maxConsecutive {
				maxConsecutive = consecutive
			}
		} else {
			consecutive = 1
			startDate = dates[i]
		}
	}

	if maxConsecutive > d.config.MaxConsecutiveDays {
		conflicts = append(conflicts, Conflict{
			Type:       ConflictConsecutive,
			Severity:   "error",
			EmployeeID: emp.ID,
			Date:       startDate,
			Message:    fmt.Sprintf("员工 %s 连续工作 %d 天，超过限制 %d 天", emp.Name, maxConsecutive, d.config.MaxConsecutiveDays),
		})
	}

	return conflicts
}

// isOverlapping 检查两个排班是否重叠
func (d *ConflictDetector) isOverlapping(a1, a2 *model.Assignment) bool {
	return a1.StartTime.Before(a2.EndTime) && a2.StartTime.Before(a1.EndTime)
}

// calculateRestHours 计算两个排班之间的休息时间
func (d *ConflictDetector) calculateRestHours(a1, a2 *model.Assignment) float64 {
	if a1.EndTime.Before(a2.StartTime) {
		return a2.StartTime.Sub(a1.EndTime).Hours()
	}
	if a2.EndTime.Before(a1.StartTime) {
		return a1.StartTime.Sub(a2.EndTime).Hours()
	}
	return -1 // 重叠
}

// calculateDailyHours 计算加上新分配后的当日工时
func (d *ConflictDetector) calculateDailyHours(newAssignment *model.Assignment, existing []*model.Assignment) float64 {
	hours := newAssignment.WorkingHours()

	for _, a := range existing {
		if a.EmployeeID == newAssignment.EmployeeID && a.Date == newAssignment.Date {
			hours += a.WorkingHours()
		}
	}

	return hours
}

// groupByEmployee 按员工分组
func groupByEmployee(assignments []*model.Assignment) map[uuid.UUID][]*model.Assignment {
	result := make(map[uuid.UUID][]*model.Assignment)
	for _, a := range assignments {
		result[a.EmployeeID] = append(result[a.EmployeeID], a)
	}
	return result
}

// isConsecutiveDateStr 检查两个日期字符串是否连续
func isConsecutiveDateStr(date1, date2 string) bool {
	t1, err1 := time.Parse("2006-01-02", date1)
	t2, err2 := time.Parse("2006-01-02", date2)
	if err1 != nil || err2 != nil {
		return false
	}

	diff := t2.Sub(t1).Hours() / 24
	return diff == 1
}

