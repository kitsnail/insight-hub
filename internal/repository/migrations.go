package repository

import (
	"database/sql"
	"fmt"
)

// migrations 数据库迁移语句
var migrations = []string{
	// 001: 创建迁移记录表（必须第一个创建）
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		applied_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 002: 创建核心表
	`CREATE TABLE IF NOT EXISTS items (
		id              TEXT PRIMARY KEY,
		type            TEXT NOT NULL,
		title           TEXT,
		content         TEXT,
		summary         TEXT,
		source          TEXT NOT NULL,
		source_url      TEXT,
		source_metadata TEXT,
		occurred_at     DATETIME,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status          TEXT NOT NULL DEFAULT 'active'
	)`,
	`CREATE INDEX IF NOT EXISTS idx_items_type ON items(type)`,
	`CREATE INDEX IF NOT EXISTS idx_items_source ON items(source)`,
	`CREATE INDEX IF NOT EXISTS idx_items_status ON items(status)`,
	`CREATE INDEX IF NOT EXISTS idx_items_occurred_at ON items(occurred_at)`,
	`CREATE INDEX IF NOT EXISTS idx_items_created_at ON items(created_at)`,

	// 002: 创建标签表
	`CREATE TABLE IF NOT EXISTS tags (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL UNIQUE,
		category    TEXT,
		color       TEXT,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)`,
	`CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category)`,

	// 003: 创建内容-标签关联表
	`CREATE TABLE IF NOT EXISTS item_tags (
		item_id     TEXT NOT NULL,
		tag_id      INTEGER NOT NULL,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (item_id, tag_id),
		FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	)`,

	// 004: 创建任务表
	`CREATE TABLE IF NOT EXISTS tasks (
		id              TEXT PRIMARY KEY,
		item_id         TEXT,
		title           TEXT NOT NULL,
		description     TEXT,
		status          TEXT NOT NULL DEFAULT 'todo',
		priority        TEXT DEFAULT 'medium',
		due_date        DATE,
		started_at      DATETIME,
		completed_at    DATETIME,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		recurrence      TEXT,
		FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE SET NULL
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date)`,

	// 005: 创建洞察表
	`CREATE TABLE IF NOT EXISTS insights (
		id              TEXT PRIMARY KEY,
		type            TEXT NOT NULL,
		title           TEXT NOT NULL,
		content         TEXT NOT NULL,
		period_start    DATE,
		period_end      DATE,
		item_ids        TEXT,
		generated_by    TEXT NOT NULL,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		status          TEXT NOT NULL DEFAULT 'active'
	)`,
	`CREATE INDEX IF NOT EXISTS idx_insights_type ON insights(type)`,
	`CREATE INDEX IF NOT EXISTS idx_insights_period ON insights(period_start, period_end)`,

	// 006: 创建数据源表
	`CREATE TABLE IF NOT EXISTS sources (
		id              TEXT PRIMARY KEY,
		name            TEXT NOT NULL UNIQUE,
		type            TEXT NOT NULL,
		config          TEXT,
		enabled         INTEGER NOT NULL DEFAULT 1,
		last_sync_at    DATETIME,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,

	// 007: 创建处理队列表
	`CREATE TABLE IF NOT EXISTS processing_queue (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		item_id         TEXT NOT NULL,
		action          TEXT NOT NULL,
		status          TEXT NOT NULL DEFAULT 'pending',
		error           TEXT,
		retry_count     INTEGER DEFAULT 0,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		processed_at    DATETIME,
		FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
	)`,
	`CREATE INDEX IF NOT EXISTS idx_processing_status ON processing_queue(status)`,

	// 008: 创建全文搜索虚拟表
	`CREATE VIRTUAL TABLE IF NOT EXISTS fts_items USING fts5(
		title,
		content,
		summary,
		content='items',
		content_rowid='rowid'
	)`,

	// 009: 创建全文搜索触发器
	`CREATE TRIGGER IF NOT EXISTS items_ai AFTER INSERT ON items BEGIN
		INSERT INTO fts_items(rowid, title, content, summary) 
		VALUES (new.rowid, new.title, new.content, new.summary);
	END`,
	`CREATE TRIGGER IF NOT EXISTS items_ad AFTER DELETE ON items BEGIN
		INSERT INTO fts_items(fts_items, rowid, title, content, summary) 
		VALUES('delete', old.rowid, old.title, old.content, old.summary);
	END`,
	`CREATE TRIGGER IF NOT EXISTS items_au AFTER UPDATE ON items BEGIN
		INSERT INTO fts_items(fts_items, rowid, title, content, summary) 
		VALUES('delete', old.rowid, old.title, old.content, old.summary);
		INSERT INTO fts_items(rowid, title, content, summary) 
		VALUES (new.rowid, new.title, new.content, new.summary);
	END`,
}

// Migrate 执行数据库迁移
func Migrate(db *sql.DB) error {
	// 先确保迁移记录表存在
	if _, err := db.Exec(migrations[0]); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 获取当前版本
	var currentVersion int
	row := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// 执行未应用的迁移（从第2个开始，第1个已执行）
	for i := currentVersion; i < len(migrations); i++ {
		// 跳过已执行的第一个迁移
		if i == 0 {
			// 记录第一个迁移
			if _, err := db.Exec("INSERT OR IGNORE INTO schema_migrations (version) VALUES (1)"); err != nil {
				return fmt.Errorf("failed to record migration 1: %w", err)
			}
			continue
		}

		if _, err := db.Exec(migrations[i]); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}

		// 记录迁移
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", i+1); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", i+1, err)
		}
	}

	return nil
}
