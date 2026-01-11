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

// Schedule 排班记录
type Schedule struct {
	ID            uuid.UUID          `json:"id"`
	OrgID         uuid.UUID          `json:"org_id"`
	Scenario      string             `json:"scenario"`
	StartDate     string             `json:"start_date"`
	EndDate       string             `json:"end_date"`
	Status        string             `json:"status"` // draft/published/archived
	TotalSlots    int                `json:"total_slots"`
	FilledSlots   int                `json:"filled_slots"`
	FillRate      float64            `json:"fill_rate"`
	Feasible      bool               `json:"feasible"`
	SoftScore     float64            `json:"soft_score"`
	GeneratedAt   time.Time          `json:"generated_at"`
	GeneratedBy   string             `json:"generated_by"` // system/manual
	Metadata      map[string]any     `json:"metadata,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// ScheduleAssignment 排班分配记录
type ScheduleAssignment struct {
	ID           uuid.UUID  `json:"id"`
	ScheduleID   uuid.UUID  `json:"schedule_id"`
	EmployeeID   uuid.UUID  `json:"employee_id"`
	EmployeeName string     `json:"employee_name"`
	ShiftID      uuid.UUID  `json:"shift_id"`
	ShiftName    string     `json:"shift_name"`
	Date         string     `json:"date"`
	StartTime    string     `json:"start_time"`
	EndTime      string     `json:"end_time"`
	Position     string     `json:"position"`
	Status       string     `json:"status"` // assigned/confirmed/cancelled
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ScheduleRepository 排班仓储接口
type ScheduleRepositoryInterface interface {
	// 排班表操作
	Create(ctx context.Context, schedule *Schedule) error
	GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error)
	Update(ctx context.Context, schedule *Schedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]*Schedule, int, error)

	// 排班分配操作
	CreateAssignment(ctx context.Context, assignment *ScheduleAssignment) error
	CreateAssignments(ctx context.Context, scheduleID uuid.UUID, assignments []*model.Assignment) error
	GetAssignments(ctx context.Context, scheduleID uuid.UUID) ([]*ScheduleAssignment, error)
	GetAssignmentsByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate string) ([]*ScheduleAssignment, error)
	DeleteAssignments(ctx context.Context, scheduleID uuid.UUID) error

	// 查询统计
	GetLatestSchedule(ctx context.Context, orgID uuid.UUID, scenario string) (*Schedule, error)
	CountByDateRange(ctx context.Context, orgID uuid.UUID, startDate, endDate string) (int, error)
}

// ScheduleRepository 排班仓储实现
type ScheduleRepository struct {
	db DB
}

// NewScheduleRepository 创建排班仓储
func NewScheduleRepository(db DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

// Create 创建排班记录
func (r *ScheduleRepository) Create(ctx context.Context, schedule *Schedule) error {
	if schedule.ID == uuid.Nil {
		schedule.ID = uuid.New()
	}
	now := time.Now()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now

	metadataJSON, _ := json.Marshal(schedule.Metadata)

	query := `
		INSERT INTO schedules (
			id, org_id, scenario, start_date, end_date, status,
			total_slots, filled_slots, fill_rate, feasible, soft_score,
			generated_at, generated_by, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		schedule.ID, schedule.OrgID, schedule.Scenario, schedule.StartDate, schedule.EndDate, schedule.Status,
		schedule.TotalSlots, schedule.FilledSlots, schedule.FillRate, schedule.Feasible, schedule.SoftScore,
		schedule.GeneratedAt, schedule.GeneratedBy, metadataJSON, schedule.CreatedAt, schedule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建排班记录失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取排班
func (r *ScheduleRepository) GetByID(ctx context.Context, id uuid.UUID) (*Schedule, error) {
	query := `
		SELECT id, org_id, scenario, start_date, end_date, status,
			total_slots, filled_slots, fill_rate, feasible, soft_score,
			generated_at, generated_by, metadata, created_at, updated_at
		FROM schedules
		WHERE id = $1
	`

	return r.scanSchedule(r.db.QueryRowContext(ctx, query, id))
}

// Update 更新排班
func (r *ScheduleRepository) Update(ctx context.Context, schedule *Schedule) error {
	schedule.UpdatedAt = time.Now()
	metadataJSON, _ := json.Marshal(schedule.Metadata)

	query := `
		UPDATE schedules SET
			status = $2, total_slots = $3, filled_slots = $4, fill_rate = $5,
			feasible = $6, soft_score = $7, metadata = $8, updated_at = $9
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		schedule.ID, schedule.Status, schedule.TotalSlots, schedule.FilledSlots, schedule.FillRate,
		schedule.Feasible, schedule.SoftScore, metadataJSON, schedule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("更新排班记录失败: %w", err)
	}

	return nil
}

// Delete 删除排班
func (r *ScheduleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// 先删除分配
	_, err := r.db.ExecContext(ctx, "DELETE FROM schedule_assignments WHERE schedule_id = $1", id)
	if err != nil {
		return fmt.Errorf("删除排班分配失败: %w", err)
	}

	// 再删除排班
	_, err = r.db.ExecContext(ctx, "DELETE FROM schedules WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("删除排班记录失败: %w", err)
	}

	return nil
}

// List 列出排班
func (r *ScheduleRepository) List(ctx context.Context, filter ListFilter) ([]*Schedule, int, error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.OrgID != nil {
		conditions = append(conditions, fmt.Sprintf("org_id = $%d", argNum))
		args = append(args, *filter.OrgID)
		argNum++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, filter.Status)
		argNum++
	}

	if filter.StartDate != "" {
		conditions = append(conditions, fmt.Sprintf("start_date >= $%d", argNum))
		args = append(args, filter.StartDate)
		argNum++
	}

	if filter.EndDate != "" {
		conditions = append(conditions, fmt.Sprintf("end_date <= $%d", argNum))
		args = append(args, filter.EndDate)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 计数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM schedules %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("统计排班数量失败: %w", err)
	}

	// 查询
	query := fmt.Sprintf(`
		SELECT id, org_id, scenario, start_date, end_date, status,
			total_slots, filled_slots, fill_rate, feasible, soft_score,
			generated_at, generated_by, metadata, created_at, updated_at
		FROM schedules %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, filter.OrderBy, filter.OrderDir, argNum, argNum+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("查询排班列表失败: %w", err)
	}
	defer rows.Close()

	var schedules []*Schedule
	for rows.Next() {
		s, err := r.scanScheduleFromRows(rows)
		if err != nil {
			return nil, 0, err
		}
		schedules = append(schedules, s)
	}

	return schedules, total, nil
}

// CreateAssignments 批量创建排班分配
func (r *ScheduleRepository) CreateAssignments(ctx context.Context, scheduleID uuid.UUID, assignments []*model.Assignment) error {
	for _, a := range assignments {
		assignment := &ScheduleAssignment{
			ID:           uuid.New(),
			ScheduleID:   scheduleID,
			EmployeeID:   a.EmployeeID,
			EmployeeName: "", // 需要从员工表查询
			ShiftID:      a.ShiftID,
			ShiftName:    "", // 需要从班次表查询
			Date:         a.Date,
			StartTime:    a.StartTime.Format("15:04"),
			EndTime:      a.EndTime.Format("15:04"),
			Position:     a.Position,
			Status:       "assigned",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := r.CreateAssignment(ctx, assignment); err != nil {
			return err
		}
	}
	return nil
}

// CreateAssignment 创建单个排班分配
func (r *ScheduleRepository) CreateAssignment(ctx context.Context, assignment *ScheduleAssignment) error {
	query := `
		INSERT INTO schedule_assignments (
			id, schedule_id, employee_id, employee_name, shift_id, shift_name,
			date, start_time, end_time, position, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		assignment.ID, assignment.ScheduleID, assignment.EmployeeID, assignment.EmployeeName,
		assignment.ShiftID, assignment.ShiftName, assignment.Date, assignment.StartTime,
		assignment.EndTime, assignment.Position, assignment.Status,
		assignment.CreatedAt, assignment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建排班分配失败: %w", err)
	}

	return nil
}

// GetAssignments 获取排班分配
func (r *ScheduleRepository) GetAssignments(ctx context.Context, scheduleID uuid.UUID) ([]*ScheduleAssignment, error) {
	query := `
		SELECT id, schedule_id, employee_id, employee_name, shift_id, shift_name,
			date, start_time, end_time, position, status, created_at, updated_at
		FROM schedule_assignments
		WHERE schedule_id = $1
		ORDER BY date, start_time
	`

	rows, err := r.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("查询排班分配失败: %w", err)
	}
	defer rows.Close()

	var assignments []*ScheduleAssignment
	for rows.Next() {
		a := &ScheduleAssignment{}
		if err := rows.Scan(
			&a.ID, &a.ScheduleID, &a.EmployeeID, &a.EmployeeName,
			&a.ShiftID, &a.ShiftName, &a.Date, &a.StartTime,
			&a.EndTime, &a.Position, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描排班分配失败: %w", err)
		}
		assignments = append(assignments, a)
	}

	return assignments, nil
}

// GetAssignmentsByEmployee 获取员工的排班分配
func (r *ScheduleRepository) GetAssignmentsByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate string) ([]*ScheduleAssignment, error) {
	query := `
		SELECT id, schedule_id, employee_id, employee_name, shift_id, shift_name,
			date, start_time, end_time, position, status, created_at, updated_at
		FROM schedule_assignments
		WHERE employee_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date, start_time
	`

	rows, err := r.db.QueryContext(ctx, query, employeeID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("查询员工排班失败: %w", err)
	}
	defer rows.Close()

	var assignments []*ScheduleAssignment
	for rows.Next() {
		a := &ScheduleAssignment{}
		if err := rows.Scan(
			&a.ID, &a.ScheduleID, &a.EmployeeID, &a.EmployeeName,
			&a.ShiftID, &a.ShiftName, &a.Date, &a.StartTime,
			&a.EndTime, &a.Position, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("扫描排班分配失败: %w", err)
		}
		assignments = append(assignments, a)
	}

	return assignments, nil
}

// DeleteAssignments 删除排班分配
func (r *ScheduleRepository) DeleteAssignments(ctx context.Context, scheduleID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM schedule_assignments WHERE schedule_id = $1", scheduleID)
	if err != nil {
		return fmt.Errorf("删除排班分配失败: %w", err)
	}
	return nil
}

// GetLatestSchedule 获取最新排班
func (r *ScheduleRepository) GetLatestSchedule(ctx context.Context, orgID uuid.UUID, scenario string) (*Schedule, error) {
	query := `
		SELECT id, org_id, scenario, start_date, end_date, status,
			total_slots, filled_slots, fill_rate, feasible, soft_score,
			generated_at, generated_by, metadata, created_at, updated_at
		FROM schedules
		WHERE org_id = $1 AND scenario = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	return r.scanSchedule(r.db.QueryRowContext(ctx, query, orgID, scenario))
}

// CountByDateRange 统计日期范围内的排班数
func (r *ScheduleRepository) CountByDateRange(ctx context.Context, orgID uuid.UUID, startDate, endDate string) (int, error) {
	query := `
		SELECT COUNT(*) FROM schedules
		WHERE org_id = $1 AND start_date >= $2 AND end_date <= $3
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, orgID, startDate, endDate).Scan(&count); err != nil {
		return 0, fmt.Errorf("统计排班数量失败: %w", err)
	}
	return count, nil
}

// scanSchedule 扫描单行排班
func (r *ScheduleRepository) scanSchedule(row *sql.Row) (*Schedule, error) {
	s := &Schedule{}
	var metadataJSON []byte

	err := row.Scan(
		&s.ID, &s.OrgID, &s.Scenario, &s.StartDate, &s.EndDate, &s.Status,
		&s.TotalSlots, &s.FilledSlots, &s.FillRate, &s.Feasible, &s.SoftScore,
		&s.GeneratedAt, &s.GeneratedBy, &metadataJSON, &s.CreatedAt, &s.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("扫描排班记录失败: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &s.Metadata)
	}

	return s, nil
}

// scanScheduleFromRows 从多行结果扫描
func (r *ScheduleRepository) scanScheduleFromRows(rows *sql.Rows) (*Schedule, error) {
	s := &Schedule{}
	var metadataJSON []byte

	err := rows.Scan(
		&s.ID, &s.OrgID, &s.Scenario, &s.StartDate, &s.EndDate, &s.Status,
		&s.TotalSlots, &s.FilledSlots, &s.FillRate, &s.Feasible, &s.SoftScore,
		&s.GeneratedAt, &s.GeneratedBy, &metadataJSON, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("扫描排班记录失败: %w", err)
	}

	if len(metadataJSON) > 0 {
		json.Unmarshal(metadataJSON, &s.Metadata)
	}

	return s, nil
}
