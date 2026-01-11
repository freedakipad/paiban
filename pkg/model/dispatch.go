// Package model 定义排班引擎的核心数据模型
package model

import (
	"time"

	"github.com/google/uuid"
)

// Customer 客户（派出服务使用）
type Customer struct {
	BaseModel
	OrgID    uuid.UUID `json:"org_id" db:"org_id"`
	Name     string    `json:"name" db:"name"`
	Code     string    `json:"code" db:"code"`
	Phone    string    `json:"phone" db:"phone"`
	Address  string    `json:"address" db:"address"`
	Location *Location `json:"location,omitempty" db:"location"`
	Type     string    `json:"type" db:"type"`           // individual/household/enterprise
	Status   string    `json:"status" db:"status"`       // active/inactive/suspended
	Level    string    `json:"level,omitempty" db:"level"` // VIP等级
	Notes    string    `json:"notes,omitempty" db:"notes"`

	// 服务需求
	ServiceNeeds    []ServiceNeed    `json:"service_needs,omitempty" db:"-"`
	Preferences     *CustomerPrefs   `json:"preferences,omitempty" db:"preferences"`
	PreferredEmpIDs []uuid.UUID      `json:"preferred_emp_ids,omitempty" db:"preferred_emp_ids"`
	BlockedEmpIDs   []uuid.UUID      `json:"blocked_emp_ids,omitempty" db:"blocked_emp_ids"`
}

// CustomerPrefs 客户偏好
type CustomerPrefs struct {
	PreferredTimes    []string          `json:"preferred_times,omitempty"`    // 偏好服务时段
	PreferredGender   string            `json:"preferred_gender,omitempty"`   // 偏好性别
	LanguageRequired  string            `json:"language_required,omitempty"`  // 语言要求
	RequireSameWorker bool              `json:"require_same_worker,omitempty"` // 要求同一服务者
	CustomPrefs       map[string]string `json:"custom,omitempty"`
}

// ServiceNeed 服务需求
type ServiceNeed struct {
	ServiceType string   `json:"service_type"` // cleaning/cooking/nursing/etc
	Frequency   string   `json:"frequency"`    // daily/weekly/biweekly
	Duration    int      `json:"duration"`     // 每次时长（分钟）
	Skills      []string `json:"skills,omitempty"`
}

// ServiceOrder 服务订单
type ServiceOrder struct {
	BaseModel
	OrgID         uuid.UUID  `json:"org_id" db:"org_id"`
	CustomerID    uuid.UUID  `json:"customer_id" db:"customer_id"`
	OrderNo       string     `json:"order_no" db:"order_no"`
	ServiceType   string     `json:"service_type" db:"service_type"`
	ServiceDate   string     `json:"service_date" db:"service_date"` // YYYY-MM-DD
	StartTime     string     `json:"start_time" db:"start_time"`     // HH:MM
	EndTime       string     `json:"end_time" db:"end_time"`         // HH:MM
	Duration      int        `json:"duration" db:"duration"`         // 分钟
	Address       string     `json:"address" db:"address"`
	Location      *Location  `json:"location,omitempty" db:"location"`
	Status        string     `json:"status" db:"status"` // pending/assigned/in_progress/completed/cancelled
	EmployeeID    *uuid.UUID `json:"employee_id,omitempty" db:"employee_id"`
	Skills        []string   `json:"skills,omitempty" db:"skills"`
	Priority      int        `json:"priority" db:"priority"`
	Notes         string     `json:"notes,omitempty" db:"notes"`
	Amount        float64    `json:"amount" db:"amount"`
	AssignedAt    *time.Time `json:"assigned_at,omitempty" db:"assigned_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// ServiceRecord 服务记录
type ServiceRecord struct {
	BaseModel
	OrderID       uuid.UUID  `json:"order_id" db:"order_id"`
	EmployeeID    uuid.UUID  `json:"employee_id" db:"employee_id"`
	CustomerID    uuid.UUID  `json:"customer_id" db:"customer_id"`
	CheckInTime   *time.Time `json:"check_in_time,omitempty" db:"check_in_time"`
	CheckOutTime  *time.Time `json:"check_out_time,omitempty" db:"check_out_time"`
	CheckInLoc    *Location  `json:"check_in_location,omitempty" db:"check_in_location"`
	CheckOutLoc   *Location  `json:"check_out_location,omitempty" db:"check_out_location"`
	ActualMinutes int        `json:"actual_minutes" db:"actual_minutes"`
	ServiceItems  []string   `json:"service_items,omitempty" db:"service_items"`
	Rating        int        `json:"rating" db:"rating"`           // 1-5
	Feedback      string     `json:"feedback,omitempty" db:"feedback"`
	Status        string     `json:"status" db:"status"`           // checked_in/checked_out/verified
	Photos        []string   `json:"photos,omitempty" db:"photos"` // 服务照片
}

// CarePlan 护理计划（长护险）
type CarePlan struct {
	BaseModel
	CustomerID      uuid.UUID   `json:"customer_id" db:"customer_id"`
	PlanNo          string      `json:"plan_no" db:"plan_no"`
	Level           int         `json:"level" db:"level"`                     // 护理等级 1-6
	StartDate       string      `json:"start_date" db:"start_date"`
	EndDate         string      `json:"end_date,omitempty" db:"end_date"`
	WeeklyHours     int         `json:"weekly_hours" db:"weekly_hours"`       // 每周服务时长
	ServiceItems    []CareItem  `json:"service_items" db:"service_items"`     // 服务项目
	Frequency       string      `json:"frequency" db:"frequency"`             // 服务频率
	PrimaryCarerID  *uuid.UUID  `json:"primary_carer_id,omitempty" db:"primary_carer_id"`
	BackupCarerIDs  []uuid.UUID `json:"backup_carer_ids,omitempty" db:"backup_carer_ids"`
	Status          string      `json:"status" db:"status"`                   // active/suspended/expired
	Notes           string      `json:"notes,omitempty" db:"notes"`
}

// CareItem 护理服务项目
type CareItem struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Frequency   int    `json:"frequency"`    // 每周次数
	Duration    int    `json:"duration"`     // 每次时长（分钟）
	RequiresCert string `json:"requires_cert,omitempty"` // 所需资质
}

// DispatchResult 派单结果
type DispatchResult struct {
	OrderID       uuid.UUID        `json:"order_id"`
	EmployeeID    uuid.UUID        `json:"employee_id"`
	EmployeeName  string           `json:"employee_name"`
	Score         float64          `json:"score"`
	Distance      float64          `json:"distance_km"`
	EstTravelTime int              `json:"est_travel_time_min"`
	MatchReasons  []string         `json:"match_reasons"`
	Conflicts     []string         `json:"conflicts,omitempty"`
	Alternatives  []AlternativeEmp `json:"alternatives,omitempty"`
}

// AlternativeEmp 备选员工
type AlternativeEmp struct {
	EmployeeID   uuid.UUID `json:"employee_id"`
	EmployeeName string    `json:"employee_name"`
	Score        float64   `json:"score"`
	Reason       string    `json:"reason"`
}

// CustomerEmployeeHistory 客户-员工服务历史
type CustomerEmployeeHistory struct {
	CustomerID    uuid.UUID `json:"customer_id" db:"customer_id"`
	EmployeeID    uuid.UUID `json:"employee_id" db:"employee_id"`
	ServiceCount  int       `json:"service_count" db:"service_count"`
	TotalMinutes  int       `json:"total_minutes" db:"total_minutes"`
	AvgRating     float64   `json:"avg_rating" db:"avg_rating"`
	LastServiceAt time.Time `json:"last_service_at" db:"last_service_at"`
	IsPrimary     bool      `json:"is_primary" db:"is_primary"`
}

// IsAssigned 检查订单是否已分配
func (o *ServiceOrder) IsAssigned() bool {
	return o.EmployeeID != nil && o.Status != "pending"
}

// NeedsDispatch 检查订单是否需要派单
func (o *ServiceOrder) NeedsDispatch() bool {
	return o.Status == "pending" && o.EmployeeID == nil
}

// IsPlanActive 检查护理计划是否有效
func (cp *CarePlan) IsPlanActive() bool {
	return cp.Status == "active"
}

