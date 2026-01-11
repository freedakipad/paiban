// Package repository 提供数据访问层
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paiban/paiban/pkg/model"
)

// ShiftRepository 班次仓储
type ShiftRepository struct {
	db DB
}

// NewShiftRepository 创建班次仓储
func NewShiftRepository(db DB) *ShiftRepository {
	return &ShiftRepository{db: db}
}

// Create 创建班次
func (r *ShiftRepository) Create(ctx context.Context, shift *model.Shift) error {
	if shift.ID == uuid.Nil {
		shift.ID = uuid.New()
	}
	now := time.Now()
	shift.CreatedAt = now
	shift.UpdatedAt = now

	query := `
		INSERT INTO shifts (
			id, org_id, name, code, description, start_time, end_time,
			duration, break_time, shift_type, color, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		shift.ID, shift.OrgID, shift.Name, shift.Code, shift.Description,
		shift.StartTime, shift.EndTime, shift.Duration, shift.BreakTime,
		shift.ShiftType, shift.Color, shift.IsActive, shift.CreatedAt, shift.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建班次失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取班次
func (r *ShiftRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Shift, error) {
	query := `
		SELECT id, org_id, name, code, description, start_time, end_time,
			duration, break_time, shift_type, color, is_active, created_at, updated_at
		FROM shifts
		WHERE id = $1 AND deleted_at IS NULL
	`

	shift := &model.Shift{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&shift.ID, &shift.OrgID, &shift.Name, &shift.Code, &shift.Description,
		&shift.StartTime, &shift.EndTime, &shift.Duration, &shift.BreakTime,
		&shift.ShiftType, &shift.Color, &shift.IsActive, &shift.CreatedAt, &shift.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询班次失败: %w", err)
	}

	return shift, nil
}

// Update 更新班次
func (r *ShiftRepository) Update(ctx context.Context, shift *model.Shift) error {
	shift.UpdatedAt = time.Now()

	query := `
		UPDATE shifts SET
			name = $2, code = $3, description = $4, start_time = $5, end_time = $6,
			duration = $7, break_time = $8, shift_type = $9, color = $10, is_active = $11, updated_at = $12
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		shift.ID, shift.Name, shift.Code, shift.Description, shift.StartTime, shift.EndTime,
		shift.Duration, shift.BreakTime, shift.ShiftType, shift.Color, shift.IsActive, shift.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("更新班次失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("班次不存在")
	}

	return nil
}

// Delete 软删除班次
func (r *ShiftRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE shifts SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("删除班次失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("班次不存在")
	}

	return nil
}

// List 查询班次列表
func (r *ShiftRepository) List(ctx context.Context, filter ListFilter) ([]*model.Shift, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.OrgID != nil {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", argIndex))
		args = append(args, *filter.OrgID)
		argIndex++
	}

	if filter.Status != "" {
		isActive := filter.Status == "active"
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, isActive)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM shifts WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询列表
	query := fmt.Sprintf(`
		SELECT id, org_id, name, code, description, start_time, end_time,
			duration, break_time, shift_type, color, is_active, created_at, updated_at
		FROM shifts
		WHERE %s
		ORDER BY start_time ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}
	defer rows.Close()

	var shifts []*model.Shift
	for rows.Next() {
		shift := &model.Shift{}
		if err := rows.Scan(
			&shift.ID, &shift.OrgID, &shift.Name, &shift.Code, &shift.Description,
			&shift.StartTime, &shift.EndTime, &shift.Duration, &shift.BreakTime,
			&shift.ShiftType, &shift.Color, &shift.IsActive, &shift.CreatedAt, &shift.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("扫描行失败: %w", err)
		}
		shifts = append(shifts, shift)
	}

	return shifts, total, nil
}

// ListActive 获取组织下所有启用的班次
func (r *ShiftRepository) ListActive(ctx context.Context, orgID uuid.UUID) ([]*model.Shift, error) {
	filter := DefaultListFilter().WithOrgID(orgID).WithStatus("active").WithLimit(100)
	shifts, _, err := r.List(ctx, filter)
	return shifts, err
}

// AssignmentRepository 排班分配仓储
type AssignmentRepository struct {
	db DB
}

// NewAssignmentRepository 创建排班分配仓储
func NewAssignmentRepository(db DB) *AssignmentRepository {
	return &AssignmentRepository{db: db}
}

// Create 创建排班分配
func (r *AssignmentRepository) Create(ctx context.Context, a *model.Assignment) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	query := `
		INSERT INTO assignments (
			id, org_id, schedule_id, employee_id, shift_id, date,
			start_time, end_time, position, status, is_overtime, is_swapped,
			original_employee_id, notes, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		a.ID, a.OrgID, a.ScheduleID, a.EmployeeID, a.ShiftID, a.Date,
		a.StartTime, a.EndTime, a.Position, a.Status, a.IsOvertime, a.IsSwapped,
		a.OriginalEmpID, a.Notes, a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建排班分配失败: %w", err)
	}

	return nil
}

// CreateBatch 批量创建排班分配
func (r *AssignmentRepository) CreateBatch(ctx context.Context, assignments []*model.Assignment) error {
	if len(assignments) == 0 {
		return nil
	}

	var values []string
	var args []interface{}
	argIndex := 1

	now := time.Now()
	for _, a := range assignments {
		if a.ID == uuid.Nil {
			a.ID = uuid.New()
		}
		a.CreatedAt = now
		a.UpdatedAt = now

		values = append(values, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5,
			argIndex+6, argIndex+7, argIndex+8, argIndex+9, argIndex+10, argIndex+11,
			argIndex+12, argIndex+13, argIndex+14, argIndex+15,
		))
		args = append(args,
			a.ID, a.OrgID, a.ScheduleID, a.EmployeeID, a.ShiftID, a.Date,
			a.StartTime, a.EndTime, a.Position, a.Status, a.IsOvertime, a.IsSwapped,
			a.OriginalEmpID, a.Notes, a.CreatedAt, a.UpdatedAt,
		)
		argIndex += 16
	}

	query := fmt.Sprintf(`
		INSERT INTO assignments (
			id, org_id, schedule_id, employee_id, shift_id, date,
			start_time, end_time, position, status, is_overtime, is_swapped,
			original_employee_id, notes, created_at, updated_at
		) VALUES %s
	`, strings.Join(values, ", "))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("批量创建排班分配失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取排班分配
func (r *AssignmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Assignment, error) {
	query := `
		SELECT id, org_id, schedule_id, employee_id, shift_id, date,
			start_time, end_time, position, status, is_overtime, is_swapped,
			original_employee_id, notes, created_at, updated_at
		FROM assignments
		WHERE id = $1 AND deleted_at IS NULL
	`

	a := &model.Assignment{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.OrgID, &a.ScheduleID, &a.EmployeeID, &a.ShiftID, &a.Date,
		&a.StartTime, &a.EndTime, &a.Position, &a.Status, &a.IsOvertime, &a.IsSwapped,
		&a.OriginalEmpID, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询排班分配失败: %w", err)
	}

	return a, nil
}

// ListBySchedule 获取排班计划的所有分配
func (r *AssignmentRepository) ListBySchedule(ctx context.Context, scheduleID uuid.UUID) ([]*model.Assignment, error) {
	query := `
		SELECT id, org_id, schedule_id, employee_id, shift_id, date,
			start_time, end_time, position, status, is_overtime, is_swapped,
			original_employee_id, notes, created_at, updated_at
		FROM assignments
		WHERE schedule_id = $1 AND deleted_at IS NULL
		ORDER BY date, start_time
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("查询排班分配失败: %w", err)
	}
	defer rows.Close()

	var assignments []*model.Assignment
	for rows.Next() {
		a := &model.Assignment{}
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.ScheduleID, &a.EmployeeID, &a.ShiftID, &a.Date,
			&a.StartTime, &a.EndTime, &a.Position, &a.Status, &a.IsOvertime, &a.IsSwapped,
			&a.OriginalEmpID, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描行失败: %w", err)
		}
		assignments = append(assignments, a)
	}

	return assignments, nil
}

// ListByEmployee 获取员工在日期范围内的排班
func (r *AssignmentRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate string) ([]*model.Assignment, error) {
	query := `
		SELECT id, org_id, schedule_id, employee_id, shift_id, date,
			start_time, end_time, position, status, is_overtime, is_swapped,
			original_employee_id, notes, created_at, updated_at
		FROM assignments
		WHERE employee_id = $1 AND date >= $2 AND date <= $3 AND deleted_at IS NULL
		ORDER BY date, start_time
	`

	rows, err := r.db.QueryContext(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("查询排班分配失败: %w", err)
	}
	defer rows.Close()

	var assignments []*model.Assignment
	for rows.Next() {
		a := &model.Assignment{}
		if err := rows.Scan(
			&a.ID, &a.OrgID, &a.ScheduleID, &a.EmployeeID, &a.ShiftID, &a.Date,
			&a.StartTime, &a.EndTime, &a.Position, &a.Status, &a.IsOvertime, &a.IsSwapped,
			&a.OriginalEmpID, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描行失败: %w", err)
		}
		assignments = append(assignments, a)
	}

	return assignments, nil
}

// Delete 软删除排班分配
func (r *AssignmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE assignments SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("删除排班分配失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("排班分配不存在")
	}

	return nil
}

// DeleteBySchedule 删除排班计划的所有分配
func (r *AssignmentRepository) DeleteBySchedule(ctx context.Context, scheduleID uuid.UUID) error {
	query := `UPDATE assignments SET deleted_at = $2 WHERE schedule_id = $1 AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, scheduleID, time.Now())
	if err != nil {
		return fmt.Errorf("删除排班分配失败: %w", err)
	}

	return nil
}

// ConstraintRepository 约束仓储
type ConstraintRepository struct {
	db DB
}

// NewConstraintRepository 创建约束仓储
func NewConstraintRepository(db DB) *ConstraintRepository {
	return &ConstraintRepository{db: db}
}

// ListByOrg 获取组织的约束配置
func (r *ConstraintRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]ConstraintConfig, error) {
	query := `
		SELECT id, name, type, category, weight, config, enabled
		FROM constraints
		WHERE org_id = $1 AND enabled = true
		ORDER BY category, weight DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("查询约束配置失败: %w", err)
	}
	defer rows.Close()

	var constraints []ConstraintConfig
	for rows.Next() {
		c := ConstraintConfig{}
		var configJSON []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Category, &c.Weight, &configJSON, &c.Enabled); err != nil {
			return nil, fmt.Errorf("扫描行失败: %w", err)
		}
		json.Unmarshal(configJSON, &c.Config)
		constraints = append(constraints, c)
	}

	return constraints, nil
}

// ConstraintConfig 约束配置
type ConstraintConfig struct {
	ID       uuid.UUID              `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Category string                 `json:"category"`
	Weight   int                    `json:"weight"`
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
}

// ScenarioTemplateRepository 场景模板仓储
type ScenarioTemplateRepository struct {
	db DB
}

// NewScenarioTemplateRepository 创建场景模板仓储
func NewScenarioTemplateRepository(db DB) *ScenarioTemplateRepository {
	return &ScenarioTemplateRepository{db: db}
}

// GetByScenario 获取场景的默认模板
func (r *ScenarioTemplateRepository) GetByScenario(ctx context.Context, scenario string) (*ScenarioTemplate, error) {
	query := `
		SELECT id, name, scenario, description, constraints, is_default, created_at
		FROM scenario_templates
		WHERE scenario = $1 AND is_default = true
		LIMIT 1
	`

	t := &ScenarioTemplate{}
	var constraintsJSON []byte
	err := r.db.QueryRowContext(ctx, query, scenario).Scan(
		&t.ID, &t.Name, &t.Scenario, &t.Description, &constraintsJSON, &t.IsDefault, &t.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询场景模板失败: %w", err)
	}

	json.Unmarshal(constraintsJSON, &t.Constraints)
	return t, nil
}

// List 获取所有场景模板
func (r *ScenarioTemplateRepository) List(ctx context.Context) ([]ScenarioTemplate, error) {
	query := `
		SELECT id, name, scenario, description, constraints, is_default, created_at
		FROM scenario_templates
		ORDER BY scenario, is_default DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询场景模板失败: %w", err)
	}
	defer rows.Close()

	var templates []ScenarioTemplate
	for rows.Next() {
		t := ScenarioTemplate{}
		var constraintsJSON []byte
		if err := rows.Scan(&t.ID, &t.Name, &t.Scenario, &t.Description, &constraintsJSON, &t.IsDefault, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("扫描行失败: %w", err)
		}
		json.Unmarshal(constraintsJSON, &t.Constraints)
		templates = append(templates, t)
	}

	return templates, nil
}

// ScenarioTemplate 场景模板
type ScenarioTemplate struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Scenario    string                 `json:"scenario"`
	Description string                 `json:"description"`
	Constraints map[string]interface{} `json:"constraints"`
	IsDefault   bool                   `json:"is_default"`
	CreatedAt   time.Time              `json:"created_at"`
}
