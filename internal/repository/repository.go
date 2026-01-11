// Package repository 提供数据访问层
package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// Repository 通用仓储接口
type Repository[T any] interface {
	Create(ctx context.Context, entity *T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, entity *T) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ListFilter) ([]*T, int, error)
}

// ListFilter 列表查询过滤器
type ListFilter struct {
	OrgID     *uuid.UUID             `json:"org_id,omitempty"`
	Status    string                 `json:"status,omitempty"`
	Search    string                 `json:"search,omitempty"`
	StartDate string                 `json:"start_date,omitempty"`
	EndDate   string                 `json:"end_date,omitempty"`
	Offset    int                    `json:"offset"`
	Limit     int                    `json:"limit"`
	OrderBy   string                 `json:"order_by,omitempty"`
	OrderDir  string                 `json:"order_dir,omitempty"` // asc/desc
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// DefaultListFilter 返回默认过滤器
func DefaultListFilter() ListFilter {
	return ListFilter{
		Offset:   0,
		Limit:    20,
		OrderBy:  "created_at",
		OrderDir: "desc",
	}
}

// WithLimit 设置限制
func (f ListFilter) WithLimit(limit int) ListFilter {
	f.Limit = limit
	return f
}

// WithOffset 设置偏移
func (f ListFilter) WithOffset(offset int) ListFilter {
	f.Offset = offset
	return f
}

// WithOrgID 设置组织ID
func (f ListFilter) WithOrgID(orgID uuid.UUID) ListFilter {
	f.OrgID = &orgID
	return f
}

// WithStatus 设置状态过滤
func (f ListFilter) WithStatus(status string) ListFilter {
	f.Status = status
	return f
}

// WithDateRange 设置日期范围
func (f ListFilter) WithDateRange(start, end string) ListFilter {
	f.StartDate = start
	f.EndDate = end
	return f
}

// DB 数据库接口
type DB interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Tx 事务接口
type Tx interface {
	DB
	Commit() error
	Rollback() error
}

// TxFunc 事务函数类型
type TxFunc func(tx Tx) error

// Scanner 行扫描接口
type Scanner interface {
	Scan(dest ...interface{}) error
}
