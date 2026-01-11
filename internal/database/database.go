// Package database 提供数据库连接和管理
package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/paiban/paiban/internal/config"
	"github.com/paiban/paiban/pkg/logger"

	_ "github.com/lib/pq" // PostgreSQL 驱动
)

// DB 数据库连接封装
type DB struct {
	*sql.DB
	cfg *config.DatabaseConfig
}

// New 创建新的数据库连接
func New(cfg *config.DatabaseConfig) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	logger.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("database", cfg.Name).
		Msg("数据库连接成功")

	return &DB{DB: db, cfg: cfg}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	if db.DB != nil {
		logger.Info().Msg("关闭数据库连接")
		return db.DB.Close()
	}
	return nil
}

// Health 健康检查
func (db *DB) Health(ctx context.Context) error {
	return db.PingContext(ctx)
}

// Transaction 执行事务
func (db *DB) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("事务回滚失败: %v (原始错误: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("事务提交失败: %w", err)
	}

	return nil
}

// Stats 返回数据库统计信息
func (db *DB) Stats() sql.DBStats {
	return db.DB.Stats()
}

// Exec 执行SQL语句
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := db.DB.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > 100*time.Millisecond {
		logger.Warn().
			Str("query", truncateQuery(query)).
			Dur("duration", duration).
			Msg("慢SQL查询")
	}

	return result, err
}

// QueryContext 执行查询
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if duration > 100*time.Millisecond {
		logger.Warn().
			Str("query", truncateQuery(query)).
			Dur("duration", duration).
			Msg("慢SQL查询")
	}

	return rows, err
}

// QueryRowContext 执行单行查询
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// truncateQuery 截断长查询
func truncateQuery(query string) string {
	if len(query) > 200 {
		return query[:200] + "..."
	}
	return query
}
