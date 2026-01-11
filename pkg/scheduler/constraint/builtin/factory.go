// Package builtin 提供内置约束实现
package builtin

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
	"github.com/paiban/paiban/pkg/scheduler/constraint"
)

// ShiftRotationPatternConstraint 倒班模式约束
// 确保员工按照指定的倒班规律轮换（如三班倒）
type ShiftRotationPatternConstraint struct {
	*BaseConstraint
	pattern      string // 倒班模式：三班倒/两班倒
	rotationDays int    // 轮换周期（天）
}

// NewShiftRotationPatternConstraint 创建倒班模式约束
func NewShiftRotationPatternConstraint(weight int, pattern string, rotationDays int) *ShiftRotationPatternConstraint {
	return &ShiftRotationPatternConstraint{
		BaseConstraint: NewBaseConstraint(
			"倒班模式",
			constraint.TypeShiftRotationPattern,
			constraint.CategoryHard,
			weight,
		),
		pattern:      pattern,
		rotationDays: rotationDays,
	}
}

// Evaluate 评估整个排班
func (c *ShiftRotationPatternConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)
		if len(assignments) < 2 {
			continue
		}

		// 按日期排序
		sorted := make([]*model.Assignment, len(assignments))
		copy(sorted, assignments)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Date < sorted[j].Date
		})

		// 检查倒班规律
		valid, details := c.checkRotationPattern(ctx, emp, sorted)
		if !valid {
			isValid = false
			for _, d := range details {
				totalPenalty += d.Penalty
				violations = append(violations, d)
			}
		}
	}

	return isValid, totalPenalty, violations
}

// checkRotationPattern 检查倒班规律
func (c *ShiftRotationPatternConstraint) checkRotationPattern(ctx *constraint.Context, emp *model.Employee, assignments []*model.Assignment) (bool, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail

	// 根据倒班模式检查
	switch c.pattern {
	case "三班倒":
		// 三班倒规则：白班->中班->夜班循环
		shiftSequence := []string{"morning", "afternoon", "night"}
		return c.checkSequencePattern(ctx, emp, assignments, shiftSequence)
	case "两班倒":
		// 两班倒规则：白班<->夜班交替
		shiftSequence := []string{"morning", "night"}
		return c.checkSequencePattern(ctx, emp, assignments, shiftSequence)
	}

	return true, violations
}

// checkSequencePattern 检查班次序列模式
func (c *ShiftRotationPatternConstraint) checkSequencePattern(ctx *constraint.Context, emp *model.Employee, assignments []*model.Assignment, _ []string) (bool, []constraint.ViolationDetail) {
	// 简化实现：主要检查不能从夜班直接到早班
	// 注：sequence参数保留用于未来完整实现顺序轮转检查
	for i := 0; i < len(assignments)-1; i++ {
		current := assignments[i]
		next := assignments[i+1]

		currentShift := ctx.GetShift(current.ShiftID)
		nextShift := ctx.GetShift(next.ShiftID)

		if currentShift == nil || nextShift == nil {
			continue
		}

		// 检查禁止的班次转换（如夜班->次日早班）
		if currentShift.ShiftType == "night" && nextShift.ShiftType == "morning" {
			// 检查是否是连续两天
			if isConsecutiveDate(current.Date, next.Date) {
				return false, []constraint.ViolationDetail{{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           next.Date,
					Message:        fmt.Sprintf("员工 %s 夜班后次日不能安排早班", emp.Name),
					Severity:       "error",
					Penalty:        c.Weight(),
				}}
			}
		}
	}

	return true, nil
}

// EvaluateAssignment 评估单个分配
func (c *ShiftRotationPatternConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	shift := ctx.GetShift(a.ShiftID)
	if shift == nil {
		return true, 0
	}

	// 检查前一天的班次
	assignments := ctx.GetEmployeeAssignments(a.EmployeeID)
	for _, existing := range assignments {
		existingShift := ctx.GetShift(existing.ShiftID)
		if existingShift == nil {
			continue
		}

		// 检查禁止的转换
		if existingShift.ShiftType == "night" && shift.ShiftType == "morning" {
			if isConsecutiveDate(existing.Date, a.Date) {
				return false, c.Weight()
			}
		}
	}

	return true, 0
}

// MaxConsecutiveNightsConstraint 最大连续夜班约束
type MaxConsecutiveNightsConstraint struct {
	*BaseConstraint
	maxNights int
}

// NewMaxConsecutiveNightsConstraint 创建最大连续夜班约束
func NewMaxConsecutiveNightsConstraint(maxNights int) *MaxConsecutiveNightsConstraint {
	return &MaxConsecutiveNightsConstraint{
		BaseConstraint: NewBaseConstraint(
			"最大连续夜班",
			constraint.TypeMaxConsecutiveNights,
			constraint.CategoryHard,
			100,
		),
		maxNights: maxNights,
	}
}

// Evaluate 评估整个排班
func (c *MaxConsecutiveNightsConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, emp := range ctx.Employees {
		assignments := ctx.GetEmployeeAssignments(emp.ID)

		// 统计连续夜班
		consecutiveNights := 0
		maxConsecutive := 0
		lastNightDate := ""

		// 按日期排序
		sorted := make([]*model.Assignment, len(assignments))
		copy(sorted, assignments)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Date < sorted[j].Date
		})

		for _, a := range sorted {
			shift := ctx.GetShift(a.ShiftID)
			if shift != nil && shift.IsNightShift() {
				if lastNightDate == "" || isConsecutiveDate(lastNightDate, a.Date) {
					consecutiveNights++
				} else {
					consecutiveNights = 1
				}
				lastNightDate = a.Date

				if consecutiveNights > maxConsecutive {
					maxConsecutive = consecutiveNights
				}
			} else {
				consecutiveNights = 0
				lastNightDate = ""
			}
		}

		if maxConsecutive > c.maxNights {
			isValid = false
			penalty := c.Weight() * (maxConsecutive - c.maxNights)
			totalPenalty += penalty

			violations = append(violations, constraint.ViolationDetail{
				ConstraintType: c.Type(),
				ConstraintName: c.Name(),
				EmployeeID:     emp.ID,
				Message:        fmt.Sprintf("员工 %s 连续夜班 %d 天，超过限制 %d 天", emp.Name, maxConsecutive, c.maxNights),
				Severity:       "error",
				Penalty:        penalty,
			})
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *MaxConsecutiveNightsConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	shift := ctx.GetShift(a.ShiftID)
	if shift == nil || !shift.IsNightShift() {
		return true, 0
	}

	// 计算加上此分配后的连续夜班数
	assignments := ctx.GetEmployeeAssignments(a.EmployeeID)
	consecutiveNights := 1

	// 往前数
	currentDate := a.Date
	for {
		prevDate := previousDate(currentDate)
		found := false
		for _, existing := range assignments {
			if existing.Date == prevDate {
				existingShift := ctx.GetShift(existing.ShiftID)
				if existingShift != nil && existingShift.IsNightShift() {
					consecutiveNights++
					currentDate = prevDate
					found = true
					break
				}
			}
		}
		if !found {
			break
		}
		if consecutiveNights > 10 {
			break
		}
	}

	if consecutiveNights > c.maxNights {
		return false, c.Weight() * (consecutiveNights - c.maxNights)
	}

	return true, 0
}

// TeamTogetherConstraint 班组完整性约束
// 确保同一班组的员工尽量安排在相同班次
type TeamTogetherConstraint struct {
	*BaseConstraint
	teams map[string][]uuid.UUID // 班组ID -> 员工ID列表
}

// NewTeamTogetherConstraint 创建班组完整性约束
func NewTeamTogetherConstraint(weight int, teams map[string][]uuid.UUID) *TeamTogetherConstraint {
	return &TeamTogetherConstraint{
		BaseConstraint: NewBaseConstraint(
			"班组完整性",
			constraint.TypeTeamTogether,
			constraint.CategorySoft,
			weight,
		),
		teams: teams,
	}
}

// Evaluate 评估整个排班
func (c *TeamTogetherConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0

	dates := getUniqueDates(ctx.Assignments)

	for _, date := range dates {
		dayAssignments := ctx.GetDateAssignments(date)

		// 按班次分组员工
		shiftEmployees := make(map[uuid.UUID][]uuid.UUID)
		for _, a := range dayAssignments {
			shiftEmployees[a.ShiftID] = append(shiftEmployees[a.ShiftID], a.EmployeeID)
		}

		// 检查每个班组是否在同一班次
		for teamID, members := range c.teams {
			if len(members) < 2 {
				continue
			}

			// 找出班组成员今天都在哪些班次
			memberShifts := make(map[uuid.UUID]uuid.UUID)
			for _, empID := range members {
				for _, a := range dayAssignments {
					if a.EmployeeID == empID {
						memberShifts[empID] = a.ShiftID
						break
					}
				}
			}

			// 检查是否在同一班次
			if len(memberShifts) >= 2 {
				shifts := make(map[uuid.UUID]int)
				for _, shiftID := range memberShifts {
					shifts[shiftID]++
				}

				if len(shifts) > 1 {
					// 班组成员分散在多个班次
					penalty := c.Weight() * (len(shifts) - 1)
					totalPenalty += penalty

					violations = append(violations, constraint.ViolationDetail{
						ConstraintType: c.Type(),
						ConstraintName: c.Name(),
						Date:           date,
						Message:        fmt.Sprintf("%s 班组 %s 成员分散在 %d 个不同班次", date, teamID, len(shifts)),
						Severity:       "warning",
						Penalty:        penalty,
					})
				}
			}
		}
	}

	return true, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *TeamTogetherConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	// 找到该员工所属班组
	var myTeam []uuid.UUID
	for _, members := range c.teams {
		for _, empID := range members {
			if empID == a.EmployeeID {
				myTeam = members
				break
			}
		}
		if myTeam != nil {
			break
		}
	}

	if myTeam == nil {
		return true, 0
	}

	// 检查班组其他成员今天的班次
	dayAssignments := ctx.GetDateAssignments(a.Date)
	for _, other := range dayAssignments {
		if other.EmployeeID == a.EmployeeID {
			continue
		}

		// 检查是否是同班组成员
		isSameTeam := false
		for _, empID := range myTeam {
			if empID == other.EmployeeID {
				isSameTeam = true
				break
			}
		}

		if isSameTeam && other.ShiftID != a.ShiftID {
			// 同班组但不同班次
			return true, c.Weight() / 2
		}
	}

	return true, 0
}

// ProductionLineCoverageConstraint 产线覆盖约束
// 确保每条产线都有足够的人员
type ProductionLineCoverageConstraint struct {
	*BaseConstraint
	lineRequirements map[string]int // 产线 -> 最少人数
}

// NewProductionLineCoverageConstraint 创建产线覆盖约束
func NewProductionLineCoverageConstraint(weight int, requirements map[string]int) *ProductionLineCoverageConstraint {
	return &ProductionLineCoverageConstraint{
		BaseConstraint: NewBaseConstraint(
			"产线覆盖",
			constraint.TypeProductionLineCoverage,
			constraint.CategoryHard,
			weight,
		),
		lineRequirements: requirements,
	}
}

// Evaluate 评估整个排班
func (c *ProductionLineCoverageConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	dates := getUniqueDates(ctx.Assignments)

	for _, date := range dates {
		dayAssignments := ctx.GetDateAssignments(date)

		// 统计每条产线的人员
		lineCoverage := make(map[string]int)
		for _, a := range dayAssignments {
			emp := ctx.GetEmployee(a.EmployeeID)
			if emp != nil && emp.Position != "" {
				// 使用岗位作为产线标识
				lineCoverage[emp.Position]++
			}
		}

		// 检查是否满足要求
		for line, minCount := range c.lineRequirements {
			actual := lineCoverage[line]
			if actual < minCount {
				isValid = false
				penalty := c.Weight() * (minCount - actual)
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					Date:           date,
					Message:        fmt.Sprintf("%s 产线 '%s' 仅有 %d 人，少于要求的 %d 人", date, line, actual, minCount),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *ProductionLineCoverageConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	return true, 0
}

// previousDate 获取前一天日期
func previousDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	prev := t.AddDate(0, 0, -1)
	return prev.Format("2006-01-02")
}

// CertificationRequiredConstraint 资质证书约束
// 确保员工具备岗位所需资质
type CertificationRequiredConstraint struct {
	*BaseConstraint
	positionCerts map[string][]string // 岗位 -> 所需证书列表
}

// NewCertificationRequiredConstraint 创建资质证书约束
func NewCertificationRequiredConstraint(weight int, certs map[string][]string) *CertificationRequiredConstraint {
	return &CertificationRequiredConstraint{
		BaseConstraint: NewBaseConstraint(
			"资质证书",
			constraint.Type("certification_required"),
			constraint.CategoryHard,
			weight,
		),
		positionCerts: certs,
	}
}

// Evaluate 评估整个排班
func (c *CertificationRequiredConstraint) Evaluate(ctx *constraint.Context) (bool, int, []constraint.ViolationDetail) {
	var violations []constraint.ViolationDetail
	totalPenalty := 0
	isValid := true

	for _, a := range ctx.Assignments {
		emp := ctx.GetEmployee(a.EmployeeID)
		if emp == nil {
			continue
		}

		position := a.Position
		if position == "" {
			position = emp.Position
		}

		requiredCerts, ok := c.positionCerts[position]
		if !ok {
			continue
		}

		for _, cert := range requiredCerts {
			if !emp.HasCertification(cert) {
				isValid = false
				penalty := c.Weight()
				totalPenalty += penalty

				violations = append(violations, constraint.ViolationDetail{
					ConstraintType: c.Type(),
					ConstraintName: c.Name(),
					EmployeeID:     emp.ID,
					Date:           a.Date,
					Message:        fmt.Sprintf("员工 %s 缺少岗位 '%s' 所需证书: %s", emp.Name, position, cert),
					Severity:       "error",
					Penalty:        penalty,
				})
			}
		}
	}

	return isValid, totalPenalty, violations
}

// EvaluateAssignment 评估单个分配
func (c *CertificationRequiredConstraint) EvaluateAssignment(ctx *constraint.Context, a *model.Assignment) (bool, int) {
	emp := ctx.GetEmployee(a.EmployeeID)
	if emp == nil {
		return false, c.Weight()
	}

	position := a.Position
	if position == "" {
		position = emp.Position
	}

	requiredCerts, ok := c.positionCerts[position]
	if !ok {
		return true, 0
	}

	for _, cert := range requiredCerts {
		if !emp.HasCertification(cert) {
			return false, c.Weight()
		}
	}

	return true, 0
}
