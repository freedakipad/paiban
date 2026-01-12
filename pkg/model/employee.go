// Package model 定义排班引擎的核心数据模型
package model

import (
	"time"

	"github.com/google/uuid"
)

// Employee 员工
type Employee struct {
	BaseModel
	OrgID    uuid.UUID `json:"org_id" db:"org_id"`
	Name     string    `json:"name" db:"name"`
	Code     string    `json:"code" db:"code"`
	Phone    string    `json:"phone,omitempty" db:"phone"`
	Email    string    `json:"email,omitempty" db:"email"`
	Status   string    `json:"status" db:"status"` // active/inactive/leave
	HireDate string    `json:"hire_date" db:"hire_date"`

	// 排班相关
	Position       string   `json:"position" db:"position"`
	Skills         []string `json:"skills" db:"skills"`
	Certifications []string `json:"certifications,omitempty" db:"certifications"`
	HourlyRate     float64  `json:"hourly_rate" db:"hourly_rate"`

	// 工作偏好
	Preferences *EmployeePreferences `json:"preferences,omitempty" db:"preferences"`

	// 每月已有班次数（前端传入，用于月度班次限制约束）
	// key: 月份 (YYYY-MM 格式), value: 该月班次数
	MonthlyShiftsCounts map[string]int `json:"monthly_shifts_counts,omitempty" db:"-"`

	// 服务区域（派出服务使用）
	ServiceArea  *ServiceArea `json:"service_area,omitempty" db:"service_area"`
	HomeLocation *Location    `json:"home_location,omitempty" db:"home_location"`
}

// EmployeePreferences 员工偏好
type EmployeePreferences struct {
	PreferredShifts   []string          `json:"preferred_shifts,omitempty"`   // 偏好班次
	AvoidShifts       []string          `json:"avoid_shifts,omitempty"`       // 避免班次
	PreferredDays     []time.Weekday    `json:"preferred_days,omitempty"`     // 偏好工作日
	AvoidDays         []time.Weekday    `json:"avoid_days,omitempty"`         // 避免工作日
	MaxHoursPerWeek   int               `json:"max_hours_per_week,omitempty"` // 期望最大周工时
	MinHoursPerWeek   int               `json:"min_hours_per_week,omitempty"` // 期望最小周工时
	CustomPreferences map[string]string `json:"custom,omitempty"`             // 自定义偏好
}

// ServiceArea 服务区域
type ServiceArea struct {
	Districts []string `json:"districts,omitempty"`  // 服务区/街道
	MaxRadius float64  `json:"max_radius,omitempty"` // 最大服务半径（公里）
	ZipCodes  []string `json:"zip_codes,omitempty"`  // 邮编列表
}

// EmployeeAvailability 员工可用性
type EmployeeAvailability struct {
	EmployeeID uuid.UUID   `json:"employee_id" db:"employee_id"`
	Date       string      `json:"date" db:"date"` // YYYY-MM-DD
	Type       string      `json:"type" db:"type"` // available/unavailable/preferred
	TimeRanges []TimeRange `json:"time_ranges,omitempty" db:"time_ranges"`
	Reason     string      `json:"reason,omitempty" db:"reason"`
}

// EmployeeContract 员工合同约束
type EmployeeContract struct {
	EmployeeID         uuid.UUID `json:"employee_id" db:"employee_id"`
	ContractType       string    `json:"contract_type" db:"contract_type"` // full_time/part_time/temp
	MinHoursPerWeek    int       `json:"min_hours_per_week" db:"min_hours_per_week"`
	MaxHoursPerWeek    int       `json:"max_hours_per_week" db:"max_hours_per_week"`
	MaxHoursPerDay     int       `json:"max_hours_per_day" db:"max_hours_per_day"`
	MaxOvertimePerWeek int       `json:"max_overtime_per_week" db:"max_overtime_per_week"`
	RestDaysPerWeek    int       `json:"rest_days_per_week" db:"rest_days_per_week"`
	ValidFrom          string    `json:"valid_from" db:"valid_from"`
	ValidTo            string    `json:"valid_to,omitempty" db:"valid_to"`
}

// IsActive 检查员工是否在职
func (e *Employee) IsActive() bool {
	return e.Status == "active"
}

// HasSkill 检查员工是否具备某技能
func (e *Employee) HasSkill(skill string) bool {
	for _, s := range e.Skills {
		if s == skill {
			return true
		}
	}
	return false
}

// HasCertification 检查员工是否具备某证书
func (e *Employee) HasCertification(cert string) bool {
	for _, c := range e.Certifications {
		if c == cert {
			return true
		}
	}
	return false
}

// CanServeLocation 检查员工是否可以服务某位置
func (e *Employee) CanServeLocation(loc Location) bool {
	if e.ServiceArea == nil || e.HomeLocation == nil {
		return true // 无限制
	}

	// 检查距离
	if e.ServiceArea.MaxRadius > 0 {
		distance := e.HomeLocation.Distance(loc)
		if distance > e.ServiceArea.MaxRadius {
			return false
		}
	}

	// 检查区域
	if len(e.ServiceArea.Districts) > 0 {
		for _, d := range e.ServiceArea.Districts {
			if d == loc.District {
				return true
			}
		}
		return false
	}

	return true
}
