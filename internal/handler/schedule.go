// Package handler 提供HTTP请求处理器
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/internal/repository"
	"github.com/paiban/paiban/pkg/errors"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
	"github.com/paiban/paiban/pkg/scheduler/constraint/builtin"
	"github.com/paiban/paiban/pkg/scheduler/solver"
)

// ScheduleHandler 排班处理器
type ScheduleHandler struct {
	scheduleRepo *repository.ScheduleRepository
	employeeRepo *repository.EmployeeRepository
	shiftRepo    *repository.ShiftRepository
}

// NewScheduleHandler 创建排班处理器
func NewScheduleHandler(
	scheduleRepo *repository.ScheduleRepository,
	employeeRepo *repository.EmployeeRepository,
	shiftRepo *repository.ShiftRepository,
) *ScheduleHandler {
	return &ScheduleHandler{
		scheduleRepo: scheduleRepo,
		employeeRepo: employeeRepo,
		shiftRepo:    shiftRepo,
	}
}

// NewScheduleHandlerWithoutDB 创建无数据库依赖的排班处理器（用于测试和简单场景）
func NewScheduleHandlerWithoutDB() *ScheduleHandler {
	return &ScheduleHandler{}
}

// GenerateRequest 排班生成请求
type GenerateRequest struct {
	OrgID        string                 `json:"org_id"`
	StartDate    string                 `json:"start_date"`
	EndDate      string                 `json:"end_date"`
	Scenario     string                 `json:"scenario,omitempty"` // restaurant/factory/housekeeping/nursing
	Employees    []EmployeeInput        `json:"employees"`
	Shifts       []ShiftInput           `json:"shifts"`
	Requirements []RequirementInput     `json:"requirements"`
	Constraints  map[string]interface{} `json:"constraints,omitempty"`
	Options      *GenerateOptions       `json:"options,omitempty"`
}

// EmployeeInput 员工输入
type EmployeeInput struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Position string   `json:"position,omitempty"`
	Skills   []string `json:"skills,omitempty"`
	Status   string   `json:"status,omitempty"`
}

// ShiftInput 班次输入
type ShiftInput struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
	Duration  int    `json:"duration"`   // 分钟
	Type      string `json:"type,omitempty"`
}

// RequirementInput 需求输入
type RequirementInput struct {
	ShiftID      string   `json:"shift_id"`
	Date         string   `json:"date"`
	Position     string   `json:"position,omitempty"`
	MinEmployees int      `json:"min_employees"`
	MaxEmployees int      `json:"max_employees,omitempty"`
	OptEmployees int      `json:"opt_employees,omitempty"`
	Skills       []string `json:"skills,omitempty"`
	Priority     int      `json:"priority,omitempty"`
}

// GenerateOptions 生成选项
type GenerateOptions struct {
	Timeout            int  `json:"timeout_seconds,omitempty"`
	OptimizationLevel  int  `json:"optimization_level,omitempty"` // 1=快速, 2=平衡, 3=最优
	RespectPreferences bool `json:"respect_preferences,omitempty"`
}

// GenerateResponse 排班生成响应
type GenerateResponse struct {
	Success     bool                    `json:"success"`
	Partial     bool                    `json:"partial,omitempty"` // 是否是部分解
	Message     string                  `json:"message,omitempty"`
	ScheduleID  string                  `json:"schedule_id,omitempty"`
	Assignments []AssignmentOutput      `json:"assignments"`
	Unfilled    []UnfilledRequirement   `json:"unfilled,omitempty"` // 未满足的需求
	Statistics  *solver.Statistics      `json:"statistics"`
	Constraints *ConstraintResultOutput `json:"constraint_result"`
	Duration    string                  `json:"duration"`
	Suggestions []StaffingSuggestion    `json:"suggestions,omitempty"` // 补员建议
}

// StaffingSuggestion 补员建议
type StaffingSuggestion struct {
	Type       string `json:"type"`        // shortage/overwork/imbalance
	Position   string `json:"position"`    // 岗位
	Date       string `json:"date"`        // 日期（可选）
	CurrentNum int    `json:"current_num"` // 当前人数
	SuggestNum int    `json:"suggest_num"` // 建议人数
	Reason     string `json:"reason"`      // 原因说明
}

// UnfilledRequirement 未满足的需求
type UnfilledRequirement struct {
	ShiftID   string `json:"shift_id"`
	ShiftName string `json:"shift_name,omitempty"`
	Date      string `json:"date"`
	Position  string `json:"position,omitempty"`
	Required  int    `json:"required"`
	Assigned  int    `json:"assigned"`
	Shortage  int    `json:"shortage"`
	Reason    string `json:"reason,omitempty"`
	StoreID   string `json:"store_id,omitempty"`
	StoreName string `json:"store_name,omitempty"`
}

// AssignmentOutput 排班输出
type AssignmentOutput struct {
	ID           string  `json:"id"`
	EmployeeID   string  `json:"employee_id"`
	EmployeeName string  `json:"employee_name,omitempty"`
	ShiftID      string  `json:"shift_id"`
	ShiftName    string  `json:"shift_name,omitempty"`
	Date         string  `json:"date"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time"`
	Position     string  `json:"position,omitempty"`
	Hours        float64 `json:"hours"`
	// 综合评分（0-100）
	Score       float64          `json:"score"`
	ScoreDetail *AssignmentScore `json:"score_detail,omitempty"`
}

// AssignmentScore 排班分配评分明细
type AssignmentScore struct {
	SkillMatch      float64  `json:"skill_match"`       // 技能匹配度 (0-100)
	Distance        float64  `json:"distance"`          // 距离评分 (0-100)
	Preference      float64  `json:"preference"`        // 偏好满足度 (0-100)
	WorkloadBalance float64  `json:"workload_balance"`  // 工时均衡 (0-100)
	Continuity      float64  `json:"continuity"`        // 连续性评分 (0-100)
	Reasons         []string `json:"reasons,omitempty"` // 评分说明
}

// ConstraintResultOutput 约束结果输出
type ConstraintResultOutput struct {
	IsValid        bool                         `json:"is_valid"`
	Score          float64                      `json:"score"`
	HardViolations []constraint.ViolationDetail `json:"hard_violations,omitempty"`
	SoftViolations []constraint.ViolationDetail `json:"soft_violations,omitempty"`
}

// Generate 生成排班
func (h *ScheduleHandler) Generate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, errors.New(errors.CodeInvalidInput, "仅支持POST方法"))
		return
	}

	// 解析请求
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "解析请求失败"))
		return
	}

	// 验证请求
	if err := validateGenerateRequest(&req); err != nil {
		respondError(w, err)
		return
	}

	// 构建排班上下文
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "无效的组织ID格式"))
		return
	}
	ctx := constraint.NewContext(orgID, req.StartDate, req.EndDate)

	// 设置员工
	employees := make([]*model.Employee, 0, len(req.Employees))
	empNameMap := make(map[uuid.UUID]string)
	empMap := make(map[uuid.UUID]*model.Employee)
	for _, e := range req.Employees {
		id, err := uuid.Parse(e.ID)
		if err != nil {
			respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "无效的员工ID格式: "+e.ID))
			return
		}
		emp := &model.Employee{
			BaseModel: model.BaseModel{ID: id},
			Name:      e.Name,
			Position:  e.Position,
			Skills:    e.Skills,
			Status:    e.Status,
		}
		if emp.Status == "" {
			emp.Status = "active"
		}
		employees = append(employees, emp)
		empNameMap[id] = e.Name
		empMap[id] = emp
	}
	ctx.SetEmployees(employees)

	// 设置班次
	shifts := make([]*model.Shift, 0, len(req.Shifts))
	shiftNameMap := make(map[uuid.UUID]string)
	for _, s := range req.Shifts {
		id, err := uuid.Parse(s.ID)
		if err != nil {
			respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "无效的班次ID格式: "+s.ID))
			return
		}
		shift := &model.Shift{
			BaseModel: model.BaseModel{ID: id},
			Name:      s.Name,
			Code:      s.Code,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Duration:  s.Duration,
			ShiftType: s.Type,
			IsActive:  true,
		}
		shifts = append(shifts, shift)
		shiftNameMap[id] = s.Name
	}
	ctx.SetShifts(shifts)

	// 设置需求
	requirements := make([]*model.ShiftRequirement, 0, len(req.Requirements))
	reqMap := make(map[string]*model.ShiftRequirement) // key: shiftID-date-position
	for _, reqItem := range req.Requirements {
		shiftID, err := uuid.Parse(reqItem.ShiftID)
		if err != nil {
			respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "无效的班次ID格式: "+reqItem.ShiftID))
			return
		}
		requirement := &model.ShiftRequirement{
			BaseModel:    model.BaseModel{ID: uuid.New()},
			ShiftID:      shiftID,
			Date:         reqItem.Date,
			Position:     reqItem.Position,
			MinEmployees: reqItem.MinEmployees,
			MaxEmployees: reqItem.MaxEmployees,
			OptEmployees: reqItem.OptEmployees,
			Skills:       reqItem.Skills,
			Priority:     reqItem.Priority,
		}
		if requirement.MaxEmployees == 0 {
			requirement.MaxEmployees = requirement.MinEmployees * 2
		}
		if requirement.Priority == 0 {
			requirement.Priority = 5
		}
		requirements = append(requirements, requirement)
		// 添加到映射
		key := fmt.Sprintf("%s-%s-%s", shiftID.String(), reqItem.Date, reqItem.Position)
		reqMap[key] = requirement
	}
	ctx.Requirements = requirements

	// 创建约束管理器并注册约束
	cm := constraint.NewManager()
	builtin.RegisterDefaultConstraints(cm, req.Constraints)

	// 创建求解器
	s := solver.NewGreedySolver(cm)

	// 设置超时上下文
	timeout := 30 * time.Second // 默认30秒超时
	if req.Options != nil && req.Options.Timeout > 0 {
		timeout = time.Duration(req.Options.Timeout) * time.Second
	}
	solveCtx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	// 执行排班
	result, err := s.Solve(solveCtx, ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			respondError(w, errors.New(errors.CodeTimeout, "排班计算超时，请尝试减少员工数量或缩短排班周期"))
			return
		}
		if err == context.Canceled {
			respondError(w, errors.New(errors.CodeInternal, "排班请求已取消"))
			return
		}
		respondError(w, errors.Wrap(err, errors.CodeInternal, "排班失败"))
		return
	}

	// 构建响应
	// 统计员工工时用于工时均衡评分
	empHours := make(map[uuid.UUID]float64)
	for _, a := range result.Assignments {
		empHours[a.EmployeeID] += a.WorkingHours()
	}
	avgHours := 0.0
	if len(empHours) > 0 {
		total := 0.0
		for _, h := range empHours {
			total += h
		}
		avgHours = total / float64(len(empHours))
	}

	// 统计员工连续工作天数
	empDays := make(map[uuid.UUID]map[string]bool)
	for _, a := range result.Assignments {
		if empDays[a.EmployeeID] == nil {
			empDays[a.EmployeeID] = make(map[string]bool)
		}
		empDays[a.EmployeeID][a.Date] = true
	}

	assignments := make([]AssignmentOutput, len(result.Assignments))
	for i, a := range result.Assignments {
		// 计算综合评分
		score, detail := calculateAssignmentScore(a, empMap[a.EmployeeID], reqMap, empHours[a.EmployeeID], avgHours, len(empDays[a.EmployeeID]))

		assignments[i] = AssignmentOutput{
			ID:           a.ID.String(),
			EmployeeID:   a.EmployeeID.String(),
			EmployeeName: empNameMap[a.EmployeeID],
			ShiftID:      a.ShiftID.String(),
			ShiftName:    shiftNameMap[a.ShiftID],
			Date:         a.Date,
			StartTime:    a.StartTime.Format("15:04"),
			EndTime:      a.EndTime.Format("15:04"),
			Position:     a.Position,
			Hours:        a.WorkingHours(),
			Score:        score,
			ScoreDetail:  detail,
		}
	}

	// 计算未满足的需求
	unfilled := calculateUnfilledRequirements(requirements, result.Assignments, shiftNameMap)
	isPartial := len(unfilled) > 0 && len(result.Assignments) > 0

	// 生成补员建议
	suggestions := generateStaffingSuggestions(unfilled, req.Employees, result.ConstraintResult)

	resp := GenerateResponse{
		Success:     result.Success,
		Partial:     isPartial,
		Message:     result.Message,
		ScheduleID:  uuid.New().String(),
		Assignments: assignments,
		Unfilled:    unfilled,
		Statistics:  result.Statistics,
		Duration:    result.Duration.String(),
		Suggestions: suggestions,
	}

	// 如果是部分解，更新消息
	if isPartial && !result.Success {
		resp.Success = true // 有部分结果就算成功
		resp.Message = "生成了部分排班方案，存在" + fmt.Sprintf("%d", len(unfilled)) + "个未满足的需求"
	}

	if result.ConstraintResult != nil {
		resp.Constraints = &ConstraintResultOutput{
			IsValid:        result.ConstraintResult.IsValid,
			Score:          result.ConstraintResult.Score,
			HardViolations: result.ConstraintResult.HardViolations,
			SoftViolations: result.ConstraintResult.SoftViolations,
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

// validateGenerateRequest 验证请求
func validateGenerateRequest(req *GenerateRequest) *errors.AppError {
	ve := &errors.ValidationErrors{}

	if req.OrgID == "" {
		ve.Add("org_id", "组织ID不能为空")
	}
	if req.StartDate == "" {
		ve.Add("start_date", "开始日期不能为空")
	}
	if req.EndDate == "" {
		ve.Add("end_date", "结束日期不能为空")
	}
	if len(req.Employees) == 0 {
		ve.Add("employees", "员工列表不能为空")
	}
	if len(req.Shifts) == 0 {
		ve.Add("shifts", "班次列表不能为空")
	}
	if len(req.Requirements) == 0 {
		ve.Add("requirements", "需求列表不能为空")
	}

	// 验证日期格式
	if req.StartDate != "" {
		if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
			ve.Add("start_date", "日期格式无效，应为YYYY-MM-DD")
		}
	}
	if req.EndDate != "" {
		if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
			ve.Add("end_date", "日期格式无效，应为YYYY-MM-DD")
		}
	}

	if ve.HasErrors() {
		return ve.ToAppError()
	}
	return nil
}

// ValidateRequest 排班验证请求
type ValidateRequest struct {
	OrgID       string                 `json:"org_id"`
	Assignments []AssignmentInput      `json:"assignments"`
	Employees   []EmployeeInput        `json:"employees"`
	Constraints map[string]interface{} `json:"constraints,omitempty"`
}

// AssignmentInput 排班输入
type AssignmentInput struct {
	EmployeeID string `json:"employee_id"`
	ShiftID    string `json:"shift_id"`
	Date       string `json:"date"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Position   string `json:"position,omitempty"`
}

// ValidateResponse 验证响应
type ValidateResponse struct {
	IsValid    bool                         `json:"is_valid"`
	Score      float64                      `json:"score"`
	Violations []constraint.ViolationDetail `json:"violations"`
}

// Validate 验证排班
func (h *ScheduleHandler) Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, errors.New(errors.CodeInvalidInput, "仅支持POST方法"))
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "解析请求失败"))
		return
	}

	// 验证组织ID
	if req.OrgID == "" {
		respondError(w, errors.New(errors.CodeInvalidInput, "组织ID不能为空"))
		return
	}

	// 构建排班上下文
	orgID, err := uuid.Parse(req.OrgID)
	if err != nil {
		respondError(w, errors.Wrap(err, errors.CodeInvalidInput, "无效的组织ID格式"))
		return
	}
	ctx := constraint.NewContext(orgID, "", "")

	// 设置员工
	employees := make([]*model.Employee, len(req.Employees))
	for i, e := range req.Employees {
		id, _ := uuid.Parse(e.ID)
		employees[i] = &model.Employee{
			BaseModel: model.BaseModel{ID: id},
			Name:      e.Name,
			Position:  e.Position,
			Skills:    e.Skills,
			Status:    "active",
		}
	}
	ctx.SetEmployees(employees)

	// 设置排班
	assignments := make([]*model.Assignment, len(req.Assignments))
	for i, a := range req.Assignments {
		empID, _ := uuid.Parse(a.EmployeeID)
		shiftID, _ := uuid.Parse(a.ShiftID)
		startTime, _ := time.Parse("2006-01-02 15:04", a.Date+" "+a.StartTime)
		endTime, _ := time.Parse("2006-01-02 15:04", a.Date+" "+a.EndTime)

		assignments[i] = &model.Assignment{
			BaseModel:  model.BaseModel{ID: uuid.New()},
			EmployeeID: empID,
			ShiftID:    shiftID,
			Date:       a.Date,
			StartTime:  startTime,
			EndTime:    endTime,
			Position:   a.Position,
		}
	}
	ctx.SetAssignments(assignments)

	// 创建约束管理器
	cm := constraint.NewManager()
	builtin.RegisterDefaultConstraints(cm, req.Constraints)

	// 评估约束
	result := cm.Evaluate(ctx)

	var violations []constraint.ViolationDetail
	violations = append(violations, result.HardViolations...)
	violations = append(violations, result.SoftViolations...)

	resp := ValidateResponse{
		IsValid:    result.IsValid,
		Score:      result.Score,
		Violations: violations,
	}

	respondJSON(w, http.StatusOK, resp)
}

// respondJSON 返回JSON响应
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError 返回错误响应
func respondError(w http.ResponseWriter, err *errors.AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   true,
		"code":    err.Code,
		"message": err.Message,
		"details": err.Details,
	})
}

// calculateUnfilledRequirements 计算未满足的需求
func calculateUnfilledRequirements(
	requirements []*model.ShiftRequirement,
	assignments []*model.Assignment,
	shiftNameMap map[uuid.UUID]string,
) []UnfilledRequirement {
	// 统计每个需求的分配数量
	assignmentCount := make(map[string]int) // key: shiftID-date-position
	for _, a := range assignments {
		key := fmt.Sprintf("%s-%s-%s", a.ShiftID.String(), a.Date, a.Position)
		assignmentCount[key]++
	}

	var unfilled []UnfilledRequirement
	for _, req := range requirements {
		key := fmt.Sprintf("%s-%s-%s", req.ShiftID.String(), req.Date, req.Position)
		assigned := assignmentCount[key]

		if assigned < req.MinEmployees {
			shortage := req.MinEmployees - assigned
			reason := "员工不足"
			if assigned == 0 {
				reason = "无可用员工"
			}

			shiftName := shiftNameMap[req.ShiftID]

			unfilled = append(unfilled, UnfilledRequirement{
				ShiftID:   req.ShiftID.String(),
				ShiftName: shiftName,
				Date:      req.Date,
				Position:  req.Position,
				Required:  req.MinEmployees,
				Assigned:  assigned,
				Shortage:  shortage,
				Reason:    reason,
			})
		}
	}

	return unfilled
}

// calculateAssignmentScore 计算单个排班分配的综合评分
func calculateAssignmentScore(
	assignment *model.Assignment,
	employee *model.Employee,
	reqMap map[string]*model.ShiftRequirement,
	empTotalHours float64,
	avgHours float64,
	empWorkDays int,
) (float64, *AssignmentScore) {
	detail := &AssignmentScore{
		SkillMatch:      100,
		Distance:        100,
		Preference:      100,
		WorkloadBalance: 100,
		Continuity:      100,
		Reasons:         []string{},
	}

	if employee == nil {
		return 50, detail
	}

	// 1. 技能匹配评分 (30%)
	key := fmt.Sprintf("%s-%s-%s", assignment.ShiftID.String(), assignment.Date, assignment.Position)
	if req, ok := reqMap[key]; ok && len(req.Skills) > 0 {
		matchedSkills := 0
		for _, reqSkill := range req.Skills {
			for _, empSkill := range employee.Skills {
				if empSkill == reqSkill {
					matchedSkills++
					break
				}
			}
		}
		if len(req.Skills) > 0 {
			detail.SkillMatch = float64(matchedSkills) / float64(len(req.Skills)) * 100
			if detail.SkillMatch >= 100 {
				detail.Reasons = append(detail.Reasons, "技能完全匹配")
			} else if detail.SkillMatch >= 50 {
				detail.Reasons = append(detail.Reasons, "技能部分匹配")
			} else {
				detail.Reasons = append(detail.Reasons, "技能匹配度低")
			}
		}
	} else {
		// 岗位匹配检查
		if employee.Position == assignment.Position {
			detail.SkillMatch = 100
			detail.Reasons = append(detail.Reasons, "岗位匹配")
		} else if assignment.Position != "" && employee.Position != "" {
			detail.SkillMatch = 60
			detail.Reasons = append(detail.Reasons, "岗位不完全匹配")
		}
	}

	// 2. 距离评分 (20%) - 如果有位置信息
	// 默认给满分，因为没有实时计算距离
	detail.Distance = 100
	if employee.HomeLocation != nil {
		detail.Reasons = append(detail.Reasons, "有住址信息")
	}

	// 3. 员工偏好评分 (20%)
	if employee.Preferences != nil {
		// 检查避免班次
		shiftCode := "" // 需要从 shiftNameMap 获取，简化处理
		for _, avoid := range employee.Preferences.AvoidShifts {
			if avoid == shiftCode || avoid == assignment.ShiftID.String() {
				detail.Preference = 30
				detail.Reasons = append(detail.Reasons, "员工避免此班次")
				break
			}
		}
		// 检查偏好班次
		for _, prefer := range employee.Preferences.PreferredShifts {
			if prefer == shiftCode || prefer == assignment.ShiftID.String() {
				detail.Preference = 100
				detail.Reasons = append(detail.Reasons, "符合员工偏好")
				break
			}
		}
	}

	// 4. 工时均衡评分 (15%)
	if avgHours > 0 {
		deviation := (empTotalHours - avgHours) / avgHours * 100
		if deviation > 20 {
			detail.WorkloadBalance = 60
			detail.Reasons = append(detail.Reasons, "工时偏高")
		} else if deviation < -20 {
			detail.WorkloadBalance = 80
			detail.Reasons = append(detail.Reasons, "工时偏低")
		} else {
			detail.WorkloadBalance = 100
			detail.Reasons = append(detail.Reasons, "工时均衡")
		}
	}

	// 5. 连续性评分 (15%)
	if empWorkDays >= 7 {
		detail.Continuity = 40
		detail.Reasons = append(detail.Reasons, "连续工作天数过多")
	} else if empWorkDays >= 6 {
		detail.Continuity = 70
		detail.Reasons = append(detail.Reasons, "接近连续工作上限")
	} else {
		detail.Continuity = 100
	}

	// 计算综合评分 (加权平均)
	score := detail.SkillMatch*0.30 +
		detail.Distance*0.20 +
		detail.Preference*0.20 +
		detail.WorkloadBalance*0.15 +
		detail.Continuity*0.15

	return score, detail
}

// generateStaffingSuggestions 生成补员建议
func generateStaffingSuggestions(unfilled []UnfilledRequirement, employees []EmployeeInput, constraintResult *constraint.Result) []StaffingSuggestion {
	var suggestions []StaffingSuggestion

	if len(unfilled) == 0 {
		return suggestions
	}

	// 按岗位统计缺口
	positionShortage := make(map[string]int)
	positionDates := make(map[string][]string)
	for _, u := range unfilled {
		positionShortage[u.Position] += u.Shortage
		positionDates[u.Position] = append(positionDates[u.Position], u.Date)
	}

	// 统计当前各岗位员工数
	positionCount := make(map[string]int)
	for _, emp := range employees {
		positionCount[emp.Position]++
	}

	// 生成补员建议
	for position, shortage := range positionShortage {
		currentNum := positionCount[position]
		// 建议增加的人数 = 缺口总数 / 排班天数 * 1.2（预留20%余量）
		uniqueDates := len(positionDates[position])
		avgShortagePerDay := float64(shortage) / float64(uniqueDates)
		suggestAdd := int(avgShortagePerDay*1.2) + 1

		suggestions = append(suggestions, StaffingSuggestion{
			Type:       "shortage",
			Position:   position,
			CurrentNum: currentNum,
			SuggestNum: currentNum + suggestAdd,
			Reason:     fmt.Sprintf("%s岗位在%d天内共缺%d个班次，建议增加%d人以满足轮换需求", position, uniqueDates, shortage, suggestAdd),
		})
	}

	// 检查是否有连续工作超限的违规
	if constraintResult != nil {
		overworkCount := 0
		for _, v := range constraintResult.HardViolations {
			if v.ConstraintType == constraint.TypeMaxConsecutiveDays {
				overworkCount++
			}
		}
		if overworkCount > 0 {
			suggestions = append(suggestions, StaffingSuggestion{
				Type:   "overwork",
				Reason: fmt.Sprintf("有%d名员工连续工作天数超限，建议增加人手以实现轮换休息", overworkCount),
			})
		}
	}

	return suggestions
}
