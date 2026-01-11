// Package solver 提供排班求解器
package solver

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/logger"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// Solver 求解器接口
type Solver interface {
	// Solve 生成排班方案
	Solve(ctx context.Context, schedCtx *constraint.Context) (*Result, error)

	// Name 返回求解器名称
	Name() string
}

// Result 求解结果
type Result struct {
	Assignments     []*model.Assignment        `json:"assignments"`
	Statistics      *Statistics                `json:"statistics"`
	ConstraintResult *constraint.Result        `json:"constraint_result"`
	Duration        time.Duration              `json:"duration"`
	Success         bool                       `json:"success"`
	Message         string                     `json:"message,omitempty"`
}

// Statistics 排班统计
type Statistics struct {
	TotalAssignments   int     `json:"total_assignments"`
	FilledRequirements int     `json:"filled_requirements"`
	TotalRequirements  int     `json:"total_requirements"`
	FillRate           float64 `json:"fill_rate"`
	TotalHours         float64 `json:"total_hours"`
	AvgHoursPerEmployee float64 `json:"avg_hours_per_employee"`
	Iterations         int     `json:"iterations"`
}

// GreedySolver 贪心求解器
type GreedySolver struct {
	constraintManager *constraint.Manager
	logger            *logger.SchedulerLogger
	maxIterations     int
}

// NewGreedySolver 创建贪心求解器
func NewGreedySolver(cm *constraint.Manager) *GreedySolver {
	return &GreedySolver{
		constraintManager: cm,
		logger:            logger.NewSchedulerLogger(),
		maxIterations:     1000,
	}
}

// Name 返回求解器名称
func (s *GreedySolver) Name() string {
	return "GreedySolver"
}

// SetMaxIterations 设置最大迭代次数
func (s *GreedySolver) SetMaxIterations(max int) {
	s.maxIterations = max
}

// Solve 使用贪心算法生成排班
func (s *GreedySolver) Solve(ctx context.Context, schedCtx *constraint.Context) (*Result, error) {
	startTime := time.Now()
	s.logger.StartSchedule(schedCtx.OrgID.String(), len(schedCtx.Employees), countDays(schedCtx.StartDate, schedCtx.EndDate))

	result := &Result{
		Assignments: make([]*model.Assignment, 0),
		Statistics:  &Statistics{},
		Success:     false,
	}

	// 检查输入
	if len(schedCtx.Employees) == 0 {
		return result, fmt.Errorf("没有可用员工")
	}
	if len(schedCtx.Requirements) == 0 {
		result.Success = true
		result.Message = "没有排班需求"
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// 按优先级和日期排序需求
	requirements := make([]*model.ShiftRequirement, len(schedCtx.Requirements))
	copy(requirements, schedCtx.Requirements)
	sort.Slice(requirements, func(i, j int) bool {
		if requirements[i].Priority != requirements[j].Priority {
			return requirements[i].Priority > requirements[j].Priority // 高优先级在前
		}
		return requirements[i].Date < requirements[j].Date // 早日期在前
	})

	// 创建员工工作量跟踪
	employeeHours := make(map[uuid.UUID]float64)
	for _, emp := range schedCtx.Employees {
		employeeHours[emp.ID] = 0
	}

	iterations := 0
	filledRequirements := 0

	// 遍历每个需求
	for _, req := range requirements {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		iterations++
		if iterations > s.maxIterations {
			break
		}

		shift := schedCtx.GetShift(req.ShiftID)
		if shift == nil {
			continue
		}

		// 计算需要分配的人数
		targetCount := req.OptEmployees
		if targetCount == 0 {
			targetCount = req.MinEmployees
		}

		assignedCount := 0

		// 获取候选员工（按工作量升序排序以保证公平）
		candidates := s.getCandidates(schedCtx, req, employeeHours)

		for _, emp := range candidates {
			if assignedCount >= targetCount {
				break
			}

			// 创建候选分配
			assignment := s.createAssignment(schedCtx, emp, req, shift)

			// 检查约束
			canAssign, reason := s.constraintManager.CanAssign(schedCtx, assignment)
			if !canAssign {
				s.logger.ConstraintViolation("分配检查", fmt.Sprintf("员工 %s: %s", emp.Name, reason))
				continue
			}

			// 添加分配
			schedCtx.AddAssignment(assignment)
			result.Assignments = append(result.Assignments, assignment)
			employeeHours[emp.ID] += assignment.WorkingHours()
			assignedCount++
		}

		if assignedCount >= req.MinEmployees {
			filledRequirements++
		}
	}

	// 评估最终结果
	result.ConstraintResult = s.constraintManager.Evaluate(schedCtx)
	result.Success = result.ConstraintResult.IsValid
	result.Duration = time.Since(startTime)

	// 统计信息
	result.Statistics.TotalAssignments = len(result.Assignments)
	result.Statistics.FilledRequirements = filledRequirements
	result.Statistics.TotalRequirements = len(requirements)
	result.Statistics.Iterations = iterations

	if len(requirements) > 0 {
		result.Statistics.FillRate = float64(filledRequirements) / float64(len(requirements)) * 100
	}

	var totalHours float64
	for _, h := range employeeHours {
		totalHours += h
	}
	result.Statistics.TotalHours = totalHours

	activeEmployees := 0
	for _, h := range employeeHours {
		if h > 0 {
			activeEmployees++
		}
	}
	if activeEmployees > 0 {
		result.Statistics.AvgHoursPerEmployee = totalHours / float64(activeEmployees)
	}

	s.logger.ScheduleComplete(schedCtx.OrgID.String(), result.Duration, result.ConstraintResult.Score)

	if !result.Success {
		result.Message = fmt.Sprintf("存在 %d 个硬约束违反", len(result.ConstraintResult.HardViolations))
	} else {
		result.Message = fmt.Sprintf("排班成功，满足率 %.1f%%", result.Statistics.FillRate)
	}

	return result, nil
}

// getCandidates 获取候选员工列表
func (s *GreedySolver) getCandidates(ctx *constraint.Context, req *model.ShiftRequirement, hours map[uuid.UUID]float64) []*model.Employee {
	var candidates []*model.Employee

	for _, emp := range ctx.Employees {
		if !emp.IsActive() {
			continue
		}

		// 检查技能匹配
		skillMatch := true
		for _, skill := range req.Skills {
			if !emp.HasSkill(skill) {
				skillMatch = false
				break
			}
		}
		if !skillMatch {
			continue
		}

		// 检查岗位匹配
		if req.Position != "" && emp.Position != req.Position {
			continue
		}

		candidates = append(candidates, emp)
	}

	// 按工作量升序排序（工作量少的优先，确保公平）
	sort.Slice(candidates, func(i, j int) bool {
		return hours[candidates[i].ID] < hours[candidates[j].ID]
	})

	return candidates
}

// createAssignment 创建排班分配
func (s *GreedySolver) createAssignment(ctx *constraint.Context, emp *model.Employee, req *model.ShiftRequirement, shift *model.Shift) *model.Assignment {
	// 解析班次时间
	date, _ := time.Parse("2006-01-02", req.Date)
	startTime := parseTimeOnDate(date, shift.StartTime)
	endTime := parseTimeOnDate(date, shift.EndTime)

	// 处理跨日班次
	if !endTime.After(startTime) {
		endTime = endTime.Add(24 * time.Hour)
	}

	return &model.Assignment{
		BaseModel:  model.BaseModel{ID: uuid.New()},
		OrgID:      ctx.OrgID,
		EmployeeID: emp.ID,
		ShiftID:    req.ShiftID,
		Date:       req.Date,
		StartTime:  startTime,
		EndTime:    endTime,
		Position:   req.Position,
		Status:     "scheduled",
	}
}

// parseTimeOnDate 在指定日期解析时间
func parseTimeOnDate(date time.Time, timeStr string) time.Time {
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return date
	}
	return time.Date(date.Year(), date.Month(), date.Day(), t.Hour(), t.Minute(), 0, 0, date.Location())
}

// countDays 计算天数
func countDays(startDate, endDate string) int {
	start, err1 := time.Parse("2006-01-02", startDate)
	end, err2 := time.Parse("2006-01-02", endDate)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

