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

// EmployeeRepository 员工仓储
type EmployeeRepository struct {
	db DB
}

// NewEmployeeRepository 创建员工仓储
func NewEmployeeRepository(db DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// Create 创建员工
func (r *EmployeeRepository) Create(ctx context.Context, emp *model.Employee) error {
	if emp.ID == uuid.Nil {
		emp.ID = uuid.New()
	}
	now := time.Now()
	emp.CreatedAt = now
	emp.UpdatedAt = now

	skillsJSON, _ := json.Marshal(emp.Skills)
	certsJSON, _ := json.Marshal(emp.Certifications)
	prefsJSON, _ := json.Marshal(emp.Preferences)
	areaJSON, _ := json.Marshal(emp.ServiceArea)
	locJSON, _ := json.Marshal(emp.HomeLocation)

	query := `
		INSERT INTO employees (
			id, org_id, name, code, phone, email, status, hire_date,
			position, skills, certifications, hourly_rate,
			preferences, service_area, home_location, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
		emp.ID, emp.OrgID, emp.Name, emp.Code, emp.Phone, emp.Email, emp.Status, emp.HireDate,
		emp.Position, skillsJSON, certsJSON, emp.HourlyRate,
		prefsJSON, areaJSON, locJSON, emp.CreatedAt, emp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建员工失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取员工
func (r *EmployeeRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Employee, error) {
	query := `
		SELECT id, org_id, name, code, phone, email, status, hire_date,
			position, skills, certifications, hourly_rate,
			preferences, service_area, home_location, created_at, updated_at
		FROM employees
		WHERE id = $1 AND deleted_at IS NULL
	`

	return r.scanEmployee(r.db.QueryRowContext(ctx, query, id))
}

// GetByCode 根据组织和工号获取员工
func (r *EmployeeRepository) GetByCode(ctx context.Context, orgID uuid.UUID, code string) (*model.Employee, error) {
	query := `
		SELECT id, org_id, name, code, phone, email, status, hire_date,
			position, skills, certifications, hourly_rate,
			preferences, service_area, home_location, created_at, updated_at
		FROM employees
		WHERE org_id = $1 AND code = $2 AND deleted_at IS NULL
	`

	return r.scanEmployee(r.db.QueryRowContext(ctx, query, orgID, code))
}

// Update 更新员工
func (r *EmployeeRepository) Update(ctx context.Context, emp *model.Employee) error {
	emp.UpdatedAt = time.Now()

	skillsJSON, _ := json.Marshal(emp.Skills)
	certsJSON, _ := json.Marshal(emp.Certifications)
	prefsJSON, _ := json.Marshal(emp.Preferences)
	areaJSON, _ := json.Marshal(emp.ServiceArea)
	locJSON, _ := json.Marshal(emp.HomeLocation)

	query := `
		UPDATE employees SET
			name = $2, code = $3, phone = $4, email = $5, status = $6,
			position = $7, skills = $8, certifications = $9, hourly_rate = $10,
			preferences = $11, service_area = $12, home_location = $13, updated_at = $14
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		emp.ID, emp.Name, emp.Code, emp.Phone, emp.Email, emp.Status,
		emp.Position, skillsJSON, certsJSON, emp.HourlyRate,
		prefsJSON, areaJSON, locJSON, emp.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("更新员工失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("员工不存在")
	}

	return nil
}

// Delete 软删除员工
func (r *EmployeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE employees SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("删除员工失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("员工不存在")
	}

	return nil
}

// List 查询员工列表
func (r *EmployeeRepository) List(ctx context.Context, filter ListFilter) ([]*model.Employee, int, error) {
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
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d OR phone ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	// 职位过滤
	if pos, ok := filter.Extra["position"].(string); ok && pos != "" {
		conditions = append(conditions, fmt.Sprintf("position = $%d", argIndex))
		args = append(args, pos)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM employees WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("查询总数失败: %w", err)
	}

	// 查询列表
	orderBy := filter.OrderBy
	if orderBy == "" {
		orderBy = "created_at"
	}
	orderDir := filter.OrderDir
	if orderDir == "" {
		orderDir = "desc"
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, name, code, phone, email, status, hire_date,
			position, skills, certifications, hourly_rate,
			preferences, service_area, home_location, created_at, updated_at
		FROM employees
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, orderDir, argIndex, argIndex+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询列表失败: %w", err)
	}
	defer rows.Close()

	var employees []*model.Employee
	for rows.Next() {
		emp, err := r.scanEmployeeRow(rows)
		if err != nil {
			return nil, 0, err
		}
		employees = append(employees, emp)
	}

	return employees, total, nil
}

// ListByIDs 根据ID列表获取员工
func (r *EmployeeRepository) ListByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Employee, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, org_id, name, code, phone, email, status, hire_date,
			position, skills, certifications, hourly_rate,
			preferences, service_area, home_location, created_at, updated_at
		FROM employees
		WHERE id IN (%s) AND deleted_at IS NULL
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询员工失败: %w", err)
	}
	defer rows.Close()

	var employees []*model.Employee
	for rows.Next() {
		emp, err := r.scanEmployeeRow(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}

	return employees, nil
}

// ListActive 获取组织下所有在职员工
func (r *EmployeeRepository) ListActive(ctx context.Context, orgID uuid.UUID) ([]*model.Employee, error) {
	filter := DefaultListFilter().WithOrgID(orgID).WithStatus("active").WithLimit(10000)
	employees, _, err := r.List(ctx, filter)
	return employees, err
}

// scanEmployee 扫描单行员工数据
func (r *EmployeeRepository) scanEmployee(row *sql.Row) (*model.Employee, error) {
	emp := &model.Employee{}
	var skillsJSON, certsJSON, prefsJSON, areaJSON, locJSON []byte

	err := row.Scan(
		&emp.ID, &emp.OrgID, &emp.Name, &emp.Code, &emp.Phone, &emp.Email, &emp.Status, &emp.HireDate,
		&emp.Position, &skillsJSON, &certsJSON, &emp.HourlyRate,
		&prefsJSON, &areaJSON, &locJSON, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("扫描员工数据失败: %w", err)
	}

	json.Unmarshal(skillsJSON, &emp.Skills)
	json.Unmarshal(certsJSON, &emp.Certifications)
	json.Unmarshal(prefsJSON, &emp.Preferences)
	json.Unmarshal(areaJSON, &emp.ServiceArea)
	json.Unmarshal(locJSON, &emp.HomeLocation)

	return emp, nil
}

// scanEmployeeRow 扫描Rows中的员工数据
func (r *EmployeeRepository) scanEmployeeRow(rows *sql.Rows) (*model.Employee, error) {
	emp := &model.Employee{}
	var skillsJSON, certsJSON, prefsJSON, areaJSON, locJSON []byte

	err := rows.Scan(
		&emp.ID, &emp.OrgID, &emp.Name, &emp.Code, &emp.Phone, &emp.Email, &emp.Status, &emp.HireDate,
		&emp.Position, &skillsJSON, &certsJSON, &emp.HourlyRate,
		&prefsJSON, &areaJSON, &locJSON, &emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("扫描员工数据失败: %w", err)
	}

	json.Unmarshal(skillsJSON, &emp.Skills)
	json.Unmarshal(certsJSON, &emp.Certifications)
	json.Unmarshal(prefsJSON, &emp.Preferences)
	json.Unmarshal(areaJSON, &emp.ServiceArea)
	json.Unmarshal(locJSON, &emp.HomeLocation)

	return emp, nil
}

