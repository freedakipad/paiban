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

// OrganizationRepository 组织仓储
type OrganizationRepository struct {
	db DB
}

// NewOrganizationRepository 创建组织仓储
func NewOrganizationRepository(db DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

// Create 创建组织
func (r *OrganizationRepository) Create(ctx context.Context, org *model.Organization) error {
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	now := time.Now()
	org.CreatedAt = now
	org.UpdatedAt = now

	settingsJSON, err := json.Marshal(org.Settings)
	if err != nil {
		return fmt.Errorf("序列化settings失败: %w", err)
	}

	query := `
		INSERT INTO organizations (id, name, code, type, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, query,
		org.ID, org.Name, org.Code, org.Type, settingsJSON, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建组织失败: %w", err)
	}

	return nil
}

// GetByID 根据ID获取组织
func (r *OrganizationRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	query := `
		SELECT id, name, code, type, settings, created_at, updated_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	org := &model.Organization{}
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID, &org.Name, &org.Code, &org.Type, &settingsJSON, &org.CreatedAt, &org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询组织失败: %w", err)
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &org.Settings); err != nil {
			return nil, fmt.Errorf("解析settings失败: %w", err)
		}
	}

	return org, nil
}

// GetByCode 根据Code获取组织
func (r *OrganizationRepository) GetByCode(ctx context.Context, code string) (*model.Organization, error) {
	query := `
		SELECT id, name, code, type, settings, created_at, updated_at
		FROM organizations
		WHERE code = $1 AND deleted_at IS NULL
	`

	org := &model.Organization{}
	var settingsJSON []byte

	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&org.ID, &org.Name, &org.Code, &org.Type, &settingsJSON, &org.CreatedAt, &org.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询组织失败: %w", err)
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &org.Settings); err != nil {
			return nil, fmt.Errorf("解析settings失败: %w", err)
		}
	}

	return org, nil
}

// Update 更新组织
func (r *OrganizationRepository) Update(ctx context.Context, org *model.Organization) error {
	org.UpdatedAt = time.Now()

	settingsJSON, err := json.Marshal(org.Settings)
	if err != nil {
		return fmt.Errorf("序列化settings失败: %w", err)
	}

	query := `
		UPDATE organizations
		SET name = $2, code = $3, type = $4, settings = $5, updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		org.ID, org.Name, org.Code, org.Type, settingsJSON, org.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("更新组织失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("组织不存在")
	}

	return nil
}

// Delete 软删除组织
func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE organizations
		SET deleted_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return fmt.Errorf("删除组织失败: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("组织不存在")
	}

	return nil
}

// List 查询组织列表
func (r *OrganizationRepository) List(ctx context.Context, filter ListFilter) ([]*model.Organization, int, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// 查询总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM organizations WHERE %s", whereClause)
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
		SELECT id, name, code, type, settings, created_at, updated_at
		FROM organizations
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

	var orgs []*model.Organization
	for rows.Next() {
		org := &model.Organization{}
		var settingsJSON []byte

		if err := rows.Scan(&org.ID, &org.Name, &org.Code, &org.Type, &settingsJSON, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("扫描行失败: %w", err)
		}

		if len(settingsJSON) > 0 {
			json.Unmarshal(settingsJSON, &org.Settings)
		}

		orgs = append(orgs, org)
	}

	return orgs, total, nil
}

