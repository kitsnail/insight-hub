package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Database 数据库封装
type Database struct {
	*sql.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(dataDir string) (*Database, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(filepath.Join(dataDir, "data"), 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(dataDir, "data", "insight-hub.db")

	// 连接数据库
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 执行迁移
	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Database{db}, nil
}

// Close 关闭数据库连接
func (db *Database) Close() error {
	return db.DB.Close()
}
