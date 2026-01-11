// Package model 定义排班引擎的核心数据模型
package model

import (
	"time"

	"github.com/google/uuid"
)

// Shift 班次定义
type Shift struct {
	BaseModel
	OrgID       uuid.UUID `json:"org_id" db:"org_id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description,omitempty" db:"description"`
	StartTime   string    `json:"start_time" db:"start_time"` // HH:MM
	EndTime     string    `json:"end_time" db:"end_time"`     // HH:MM
	Duration    int       `json:"duration" db:"duration"`     // 分钟
	BreakTime   int       `json:"break_time" db:"break_time"` // 休息时间（分钟）
	ShiftType   string    `json:"shift_type" db:"shift_type"` // morning/afternoon/evening/night/split
	Color       string    `json:"color,omitempty" db:"color"` // 颜色标识
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// ShiftRequirement 班次需求
type ShiftRequirement struct {
	BaseModel
	OrgID        uuid.UUID `json:"org_id" db:"org_id"`
	ShiftID      uuid.UUID `json:"shift_id" db:"shift_id"`
	Date         string    `json:"date" db:"date"` // YYYY-MM-DD
	Position     string    `json:"position,omitempty" db:"position"`
	MinEmployees int       `json:"min_employees" db:"min_employees"`
	MaxEmployees int       `json:"max_employees" db:"max_employees"`
	OptEmployees int       `json:"opt_employees" db:"opt_employees"` // 最优人数
	Skills       []string  `json:"skills,omitempty" db:"skills"`
	Priority     int       `json:"priority" db:"priority"` // 优先级 1-10
}

// Assignment 排班分配
type Assignment struct {
	BaseModel
	OrgID         uuid.UUID  `json:"org_id" db:"org_id"`
	ScheduleID    uuid.UUID  `json:"schedule_id" db:"schedule_id"`
	EmployeeID    uuid.UUID  `json:"employee_id" db:"employee_id"`
	ShiftID       uuid.UUID  `json:"shift_id" db:"shift_id"`
	Date          string     `json:"date" db:"date"`
	StartTime     time.Time  `json:"start_time" db:"start_time"`
	EndTime       time.Time  `json:"end_time" db:"end_time"`
	Position      string     `json:"position,omitempty" db:"position"`
	Status        string     `json:"status" db:"status"` // scheduled/confirmed/completed/cancelled
	IsOvertime    bool       `json:"is_overtime" db:"is_overtime"`
	IsSwapped     bool       `json:"is_swapped" db:"is_swapped"`
	OriginalEmpID *uuid.UUID `json:"original_employee_id,omitempty" db:"original_employee_id"`
	Notes         string     `json:"notes,omitempty" db:"notes"`
}

// Schedule 排班计划
type Schedule struct {
	BaseModel
	OrgID       uuid.UUID      `json:"org_id" db:"org_id"`
	Name        string         `json:"name" db:"name"`
	StartDate   string         `json:"start_date" db:"start_date"`
	EndDate     string         `json:"end_date" db:"end_date"`
	Status      string         `json:"status" db:"status"` // draft/published/archived
	Version     int            `json:"version" db:"version"`
	CreatedBy   *uuid.UUID     `json:"created_by,omitempty" db:"created_by"`
	PublishedAt *time.Time     `json:"published_at,omitempty" db:"published_at"`
	Assignments []Assignment   `json:"assignments,omitempty" db:"-"`
	Statistics  *ScheduleStats `json:"statistics,omitempty" db:"-"`
}

// ScheduleStats 排班统计
type ScheduleStats struct {
	TotalAssignments int     `json:"total_assignments"`
	TotalEmployees   int     `json:"total_employees"`
	TotalHours       float64 `json:"total_hours"`
	OvertimeHours    float64 `json:"overtime_hours"`
	UnfilledShifts   int     `json:"unfilled_shifts"`
	ConstraintScore  float64 `json:"constraint_score"` // 约束满足率
	FairnessScore    float64 `json:"fairness_score"`   // 公平性得分
	PreferenceScore  float64 `json:"preference_score"` // 偏好满足率
}

// SwapRequest 换班请求
type SwapRequest struct {
	BaseModel
	OrgID            uuid.UUID  `json:"org_id" db:"org_id"`
	RequestorID      uuid.UUID  `json:"requestor_id" db:"requestor_id"`
	TargetID         *uuid.UUID `json:"target_id,omitempty" db:"target_id"`
	SourceAssignment uuid.UUID  `json:"source_assignment" db:"source_assignment"`
	TargetAssignment *uuid.UUID `json:"target_assignment,omitempty" db:"target_assignment"`
	Status           string     `json:"status" db:"status"` // pending/approved/rejected/cancelled
	Reason           string     `json:"reason,omitempty" db:"reason"`
	ReviewedBy       *uuid.UUID `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty" db:"reviewed_at"`
	ReviewNote       string     `json:"review_note,omitempty" db:"review_note"`
}

// WorkingHours 计算工作时长（小时）
func (a *Assignment) WorkingHours() float64 {
	return a.EndTime.Sub(a.StartTime).Hours()
}

// IsOnDate 检查分配是否在指定日期
func (a *Assignment) IsOnDate(date string) bool {
	return a.Date == date
}

// DurationHours 返回班次时长（小时）
func (s *Shift) DurationHours() float64 {
	return float64(s.Duration-s.BreakTime) / 60.0
}

// IsNightShift 检查是否为夜班
func (s *Shift) IsNightShift() bool {
	return s.ShiftType == "night"
}

// IsSplitShift 检查是否为两头班
func (s *Shift) IsSplitShift() bool {
	return s.ShiftType == "split"
}
