// Package constraint 定义约束接口和管理器
package constraint

import (
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

// Type 约束类型标识
type Type string

const (
	// 硬约束类型
	TypeMaxHoursPerDay         Type = "max_hours_per_day"
	TypeMaxHoursPerWeek        Type = "max_hours_per_week"
	TypeMinRestBetweenShifts   Type = "min_rest_between_shifts"
	TypeMaxConsecutiveDays     Type = "max_consecutive_days"
	TypeMaxShiftsPerDay        Type = "max_shifts_per_day"
	TypeSkillRequired          Type = "skill_required"
	TypeProductionLineCoverage Type = "production_line_coverage"
	TypeShiftRotationPattern   Type = "shift_rotation_pattern"
	TypeMaxConsecutiveNights   Type = "max_consecutive_night_shifts"
	TypeServiceAreaMatch       Type = "service_area_match"
	TypeTravelTimeBuffer       Type = "travel_time_buffer"
	TypeMaxOrdersPerDay        Type = "max_orders_per_day"
	TypeCarePlanCompliance     Type = "care_plan_compliance"
	TypeCertificationLevel     Type = "certification_level"

	// 软约束类型
	TypeEmployeePreference     Type = "employee_preference"
	TypeWorkloadBalance        Type = "workload_balance"
	TypeMinimizeOvertime       Type = "minimize_overtime"
	TypePeakHoursCoverage      Type = "peak_hours_coverage"
	TypeTeamTogether           Type = "team_together"
	TypeCustomerPreference     Type = "customer_preference"
	TypeMinimizeTravelDistance Type = "minimize_travel_distance"
	TypeServiceContinuity      Type = "service_continuity"
	TypeCaregiverContinuity    Type = "caregiver_continuity"
)

// Category 约束类别
type Category string

const (
	CategoryHard Category = "hard" // 硬约束（必须满足）
	CategorySoft Category = "soft" // 软约束（尽量满足）
)

// Constraint 约束接口
type Constraint interface {
	// Name 返回约束名称
	Name() string

	// Type 返回约束类型
	Type() Type

	// Category 返回约束类别
	Category() Category

	// Weight 返回约束权重 (1-100)
	Weight() int

	// Evaluate 评估整个排班方案
	// 返回：是否满足、惩罚值、违反详情
	Evaluate(ctx *Context) (valid bool, penalty int, details []ViolationDetail)

	// EvaluateAssignment 评估单个分配
	// 返回：是否满足、惩罚值
	EvaluateAssignment(ctx *Context, assignment *model.Assignment) (valid bool, penalty int)
}

// ViolationDetail 约束违反详情
type ViolationDetail struct {
	ConstraintType Type      `json:"constraint_type"`
	ConstraintName string    `json:"constraint_name"`
	EmployeeID     uuid.UUID `json:"employee_id,omitempty"`
	Date           string    `json:"date,omitempty"`
	Message        string    `json:"message"`
	Severity       string    `json:"severity"` // error/warning
	Penalty        int       `json:"penalty"`
}

// Context 排班上下文
type Context struct {
	// 输入数据
	OrgID        uuid.UUID                 `json:"org_id"`
	StartDate    string                    `json:"start_date"`
	EndDate      string                    `json:"end_date"`
	Employees    []*model.Employee         `json:"employees"`
	Shifts       []*model.Shift            `json:"shifts"`
	Requirements []*model.ShiftRequirement `json:"requirements"`

	// 当前排班结果
	Assignments []*model.Assignment `json:"assignments"`

	// 索引缓存
	employeeMap       map[uuid.UUID]*model.Employee
	shiftMap          map[uuid.UUID]*model.Shift
	assignmentsByEmp  map[uuid.UUID][]*model.Assignment
	assignmentsByDate map[string][]*model.Assignment

	// 额外配置
	Config map[string]interface{} `json:"config,omitempty"`
}

// NewContext 创建新的排班上下文
func NewContext(orgID uuid.UUID, startDate, endDate string) *Context {
	return &Context{
		OrgID:             orgID,
		StartDate:         startDate,
		EndDate:           endDate,
		Employees:         make([]*model.Employee, 0),
		Shifts:            make([]*model.Shift, 0),
		Requirements:      make([]*model.ShiftRequirement, 0),
		Assignments:       make([]*model.Assignment, 0),
		employeeMap:       make(map[uuid.UUID]*model.Employee),
		shiftMap:          make(map[uuid.UUID]*model.Shift),
		assignmentsByEmp:  make(map[uuid.UUID][]*model.Assignment),
		assignmentsByDate: make(map[string][]*model.Assignment),
		Config:            make(map[string]interface{}),
	}
}

// SetEmployees 设置员工列表
func (c *Context) SetEmployees(employees []*model.Employee) {
	c.Employees = employees
	c.employeeMap = make(map[uuid.UUID]*model.Employee)
	for _, e := range employees {
		c.employeeMap[e.ID] = e
	}
}

// SetShifts 设置班次列表
func (c *Context) SetShifts(shifts []*model.Shift) {
	c.Shifts = shifts
	c.shiftMap = make(map[uuid.UUID]*model.Shift)
	for _, s := range shifts {
		c.shiftMap[s.ID] = s
	}
}

// SetAssignments 设置排班分配
func (c *Context) SetAssignments(assignments []*model.Assignment) {
	c.Assignments = assignments
	c.rebuildAssignmentIndexes()
}

// AddAssignment 添加排班分配
func (c *Context) AddAssignment(a *model.Assignment) {
	c.Assignments = append(c.Assignments, a)
	c.assignmentsByEmp[a.EmployeeID] = append(c.assignmentsByEmp[a.EmployeeID], a)
	c.assignmentsByDate[a.Date] = append(c.assignmentsByDate[a.Date], a)
}

// RemoveAssignment 移除排班分配
func (c *Context) RemoveAssignment(id uuid.UUID) {
	for i, a := range c.Assignments {
		if a.ID == id {
			c.Assignments = append(c.Assignments[:i], c.Assignments[i+1:]...)
			break
		}
	}
	c.rebuildAssignmentIndexes()
}

// rebuildAssignmentIndexes 重建分配索引
func (c *Context) rebuildAssignmentIndexes() {
	c.assignmentsByEmp = make(map[uuid.UUID][]*model.Assignment)
	c.assignmentsByDate = make(map[string][]*model.Assignment)
	for _, a := range c.Assignments {
		c.assignmentsByEmp[a.EmployeeID] = append(c.assignmentsByEmp[a.EmployeeID], a)
		c.assignmentsByDate[a.Date] = append(c.assignmentsByDate[a.Date], a)
	}
}

// GetEmployee 获取员工
func (c *Context) GetEmployee(id uuid.UUID) *model.Employee {
	return c.employeeMap[id]
}

// GetShift 获取班次
func (c *Context) GetShift(id uuid.UUID) *model.Shift {
	return c.shiftMap[id]
}

// GetEmployeeAssignments 获取员工的所有排班
func (c *Context) GetEmployeeAssignments(empID uuid.UUID) []*model.Assignment {
	return c.assignmentsByEmp[empID]
}

// GetDateAssignments 获取某日期的所有排班
func (c *Context) GetDateAssignments(date string) []*model.Assignment {
	return c.assignmentsByDate[date]
}

// GetEmployeeHoursOnDate 获取员工某天的工作时长
func (c *Context) GetEmployeeHoursOnDate(empID uuid.UUID, date string) float64 {
	var hours float64
	for _, a := range c.assignmentsByEmp[empID] {
		if a.Date == date {
			hours += a.WorkingHours()
		}
	}
	return hours
}

// GetEmployeeHoursInRange 获取员工在日期范围内的工作时长
func (c *Context) GetEmployeeHoursInRange(empID uuid.UUID, startDate, endDate string) float64 {
	var hours float64
	for _, a := range c.assignmentsByEmp[empID] {
		if a.Date >= startDate && a.Date <= endDate {
			hours += a.WorkingHours()
		}
	}
	return hours
}

// GetEmployeeConsecutiveDays 获取员工在指定日期前后的连续工作天数
// 返回：如果在该日期分配，会形成的最大连续工作天数
func (c *Context) GetEmployeeConsecutiveDays(empID uuid.UUID, targetDate string) int {
	// 获取员工排班的日期
	dates := make(map[string]bool)
	for _, a := range c.assignmentsByEmp[empID] {
		dates[a.Date] = true
	}

	// 往前数连续工作天数（不包括目标日期）
	countBefore := 0
	currentDate := previousDate(targetDate)
	for dates[currentDate] {
		countBefore++
		currentDate = previousDate(currentDate)
		if countBefore > 30 { // 防止无限循环
			break
		}
	}

	// 往后数连续工作天数（不包括目标日期）
	countAfter := 0
	currentDate = nextDate(targetDate)
	for dates[currentDate] {
		countAfter++
		currentDate = nextDate(currentDate)
		if countAfter > 30 { // 防止无限循环
			break
		}
	}

	// 返回：前面连续天数 + 后面连续天数（如果分配目标日期，总连续天数 = countBefore + 1 + countAfter）
	// 但这里只返回前面的连续天数，因为调用方会 +1
	// 为了正确计算，我们返回 countBefore + countAfter，让调用方 +1 后得到正确的总连续天数
	return countBefore + countAfter
}

// previousDate 获取前一天日期
func previousDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	return t.AddDate(0, 0, -1).Format("2006-01-02")
}

// nextDate 获取后一天日期
func nextDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

// Result 约束评估结果
type Result struct {
	IsValid        bool              `json:"is_valid"`
	TotalPenalty   int               `json:"total_penalty"`
	HardViolations []ViolationDetail `json:"hard_violations"`
	SoftViolations []ViolationDetail `json:"soft_violations"`
	Score          float64           `json:"score"` // 0-100
}

// CalculateScore 计算约束满足度得分
func (r *Result) CalculateScore(maxPenalty int) {
	if maxPenalty == 0 {
		r.Score = 100.0
		return
	}
	r.Score = 100.0 * float64(maxPenalty-r.TotalPenalty) / float64(maxPenalty)
	if r.Score < 0 {
		r.Score = 0
	}
}
