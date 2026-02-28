package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitsnail/insight-hub/internal/config"
)

// PostgresDB PostgreSQL 数据库封装
type PostgresDB struct {
	*pgxpool.Pool
}

// NewPostgresDB 创建 PostgreSQL 连接池
func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	// 构建连接字符串
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	// 解析连接池配置
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// 设置连接池参数
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns

	if cfg.MaxConnLifetime != "" {
		if d, err := time.ParseDuration(cfg.MaxConnLifetime); err == nil {
			poolConfig.MaxConnLifetime = d
		}
	}

	if cfg.MaxConnIdleTime != "" {
		if d, err := time.ParseDuration(cfg.MaxConnIdleTime); err == nil {
			poolConfig.MaxConnIdleTime = d
		}
	}

	// 创建连接池
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 测试连接
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 执行迁移
	if err := MigratePostgres(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &PostgresDB{pool}, nil
}

// Close 关闭连接池
func (db *PostgresDB) Close() {
	db.Pool.Close()
}
