package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// postgresMigrations PostgreSQL 数据库迁移语句
var postgresMigrations = []string{
	// 001: 创建迁移记录表
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version     INTEGER PRIMARY KEY,
		applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	// 002: 启用扩展
	`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
	`CREATE EXTENSION IF NOT EXISTS "pg_trgm"`,

	// 003: 创建核心表
	`CREATE TABLE IF NOT EXISTS items (
		id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		type            TEXT NOT NULL,
		title           TEXT,
		content         TEXT,
		summary         TEXT,
		source          TEXT NOT NULL,
		source_url      TEXT,
		source_metadata JSONB,
		occurred_at     TIMESTAMPTZ,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		status          TEXT NOT NULL DEFAULT 'active'
	)`,
	`CREATE INDEX IF NOT EXISTS idx_items_type ON items(type)`,
	`CREATE INDEX IF NOT EXISTS idx_items_source ON items(source)`,
	`CREATE INDEX IF NOT EXISTS idx_items_status ON items(status)`,
	`CREATE INDEX IF NOT EXISTS idx_items_occurred_at ON items(occurred_at)`,
	`CREATE INDEX IF NOT EXISTS idx_items_created_at ON items(created_at)`,

	// 004: 创建标签表
	`CREATE TABLE IF NOT EXISTS tags (
		id          SERIAL PRIMARY KEY,
		name        TEXT NOT NULL UNIQUE,
		category    TEXT,
		color       TEXT,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name)`,
	`CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category)`,

	// 005: 创建内容-标签关联表
	`CREATE TABLE IF NOT EXISTS item_tags (
		item_id     UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
		tag_id      INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (item_id, tag_id)
	)`,

	// 006: 创建任务表
	`CREATE TABLE IF NOT EXISTS tasks (
		id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		item_id         UUID REFERENCES items(id) ON DELETE SET NULL,
		title           TEXT NOT NULL,
		description     TEXT,
		status          TEXT NOT NULL DEFAULT 'todo',
		priority        TEXT DEFAULT 'medium',
		due_date        DATE,
		started_at      TIMESTAMPTZ,
		completed_at    TIMESTAMPTZ,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		recurrence      TEXT
	)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority)`,
	`CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date)`,

	// 007: 创建洞察表
	`CREATE TABLE IF NOT EXISTS insights (
		id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		type            TEXT NOT NULL,
		title           TEXT NOT NULL,
		content         TEXT NOT NULL,
		period_start    DATE,
		period_end      DATE,
		item_ids        JSONB,
		generated_by    TEXT NOT NULL,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		status          TEXT NOT NULL DEFAULT 'active'
	)`,
	`CREATE INDEX IF NOT EXISTS idx_insights_type ON insights(type)`,
	`CREATE INDEX IF NOT EXISTS idx_insights_period ON insights(period_start, period_end)`,

	// 008: 创建数据源表
	`CREATE TABLE IF NOT EXISTS sources (
		id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name            TEXT NOT NULL UNIQUE,
		type            TEXT NOT NULL,
		config          JSONB,
		enabled         BOOLEAN NOT NULL DEFAULT TRUE,
		last_sync_at    TIMESTAMPTZ,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`,

	// 009: 创建处理队列表
	`CREATE TABLE IF NOT EXISTS processing_queue (
		id              SERIAL PRIMARY KEY,
		item_id         UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
		action          TEXT NOT NULL,
		status          TEXT NOT NULL DEFAULT 'pending',
		error           TEXT,
		retry_count     INTEGER DEFAULT 0,
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		processed_at    TIMESTAMPTZ
	)`,
	`CREATE INDEX IF NOT EXISTS idx_processing_status ON processing_queue(status)`,

	// 010: 创建全文搜索索引 (使用 PostgreSQL 内置全文搜索)
	`CREATE INDEX IF NOT EXISTS idx_items_title_trgm ON items USING gin(title gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_items_content_trgm ON items USING gin(content gin_trgm_ops)`,
	`CREATE INDEX IF NOT EXISTS idx_items_summary_trgm ON items USING gin(summary gin_trgm_ops)`,

	// 011: 创建去重表
	`CREATE TABLE IF NOT EXISTS dedup (
		fingerprint TEXT PRIMARY KEY,
		item_id     UUID NOT NULL REFERENCES items(id) ON DELETE CASCADE,
		source      TEXT NOT NULL,
		created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		expires_at  TIMESTAMPTZ
	)`,
	`CREATE INDEX IF NOT EXISTS idx_dedup_expires ON dedup(expires_at) WHERE expires_at IS NOT NULL`,

	// ============================================
	// v2 迁移: 数据模型重构 (2026-03-01)
	// ============================================

	// 012: 添加 agent 字段
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS agent VARCHAR(32) NOT NULL DEFAULT 'unknown'`,
	`CREATE INDEX IF NOT EXISTS idx_items_agent ON items(agent)`,

	// 013: 删除 type 约束（v3 开放类型）
	`ALTER TABLE items DROP CONSTRAINT IF EXISTS items_type_check`,

	// 014: 更新现有数据的 type
	`UPDATE items SET type = 'news' WHERE type = 'bookmark'`,
	`UPDATE items SET type = 'insight' WHERE type = 'note'`,
	`UPDATE items SET type = 'log' WHERE type = 'chat'`,
	`UPDATE items SET type = 'log' WHERE type = 'code'`,
	`UPDATE items SET type = 'log' WHERE type = 'task'`,

	// 015: 更新现有数据的 agent
	`UPDATE items SET agent = 'news-agent' WHERE source LIKE '%news%' OR source LIKE '%rss%' OR source LIKE '%feed%'`,
	`UPDATE items SET agent = 'tech-doc-agent' WHERE source LIKE '%github%' OR source LIKE '%doc%' OR source LIKE '%tech%'`,

	// 016: 创建新索引
	`CREATE INDEX IF NOT EXISTS idx_items_metadata ON items USING GIN (source_metadata)`,
	`CREATE INDEX IF NOT EXISTS idx_items_fulltext ON items 
	  USING GIN (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, '')))`,

	// 017: 创建视图
	`CREATE OR REPLACE VIEW active_trackings AS
	SELECT 
	    id,
	    agent,
	    title,
	    source_metadata->>'track_id' AS track_id,
	    source_metadata->>'track_type' AS track_type,
	    source_metadata->>'status' AS status,
	    source_metadata->>'last_check' AS last_check,
	    occurred_at,
	    updated_at
	FROM items
	WHERE type = 'tracking'
	  AND source_metadata->>'status' = 'active'
	ORDER BY updated_at DESC`,

	`CREATE OR REPLACE VIEW recent_briefs AS
	SELECT 
	    id,
	    agent,
	    title,
	    occurred_at,
	    source_metadata->>'sent_to' AS sent_to,
	    source_metadata->>'sent_at' AS sent_at
	FROM items
	WHERE type = 'brief'
	ORDER BY occurred_at DESC
	LIMIT 30`,

	// 018: 更新去重表
	`ALTER TABLE dedup ADD COLUMN IF NOT EXISTS agent VARCHAR(32)`,
	`CREATE INDEX IF NOT EXISTS idx_dedup_agent ON dedup(agent)`,

	// ============================================
	// v3 迁移: 个人数据中枢 (2026-03-01)
	// ============================================

	// 019: 添加 v3 字段
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS category VARCHAR(64)`,
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS source_system VARCHAR(64)`,
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS source_agent VARCHAR(64)`,
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS metadata JSONB`,
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0`,
	`ALTER TABLE items ADD COLUMN IF NOT EXISTS search_vector TSVECTOR`,

	// 020: 迁移数据
	`UPDATE items SET 
		source_system = COALESCE(agent, 'unknown'),
		source_agent = agent,
		metadata = COALESCE(source_metadata, '{}'::jsonb)
	WHERE source_system IS NULL`,

	// 021: 创建 v3 索引
	`CREATE INDEX IF NOT EXISTS idx_items_category ON items(category)`,
	`CREATE INDEX IF NOT EXISTS idx_items_source_system ON items(source_system)`,
	`CREATE INDEX IF NOT EXISTS idx_items_priority ON items(priority DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_items_metadata_v3 ON items USING GIN (metadata)`,
	`CREATE INDEX IF NOT EXISTS idx_items_search_vector_v3 ON items USING GIN (search_vector)`,

	// 022: 更新搜索向量
	`UPDATE items SET search_vector = to_tsvector('english',
		coalesce(title, '') || ' ' ||
		coalesce(content, '') || ' ' ||
		coalesce(summary, '')
	)`,

	// 023: 创建来源系统注册表
	`CREATE TABLE IF NOT EXISTS source_systems (
		name        VARCHAR(64) PRIMARY KEY,
		description TEXT,
		config      JSONB DEFAULT '{}',
		created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	)`,

	// 024: 初始化来源系统
	`INSERT INTO source_systems (name, description) VALUES
		('openclaw', 'openclaw AI system'),
		('picoclaw', 'picoclaw AI system'),
		('opencode', 'opencode development tool'),
		('claude-code', 'Claude Code assistant'),
		('codex', 'OpenAI Codex'),
		('manual', 'Manual entry'),
		('web', 'Web scraper')
	ON CONFLICT (name) DO NOTHING`,

	// 025: 创建统计视图
	`CREATE OR REPLACE VIEW stats_by_source AS
	SELECT 
		source_system,
		type,
		COUNT(*) AS count,
		MAX(created_at) AS last_created
	FROM items
	WHERE status = 'active'
	GROUP BY source_system, type
	ORDER BY source_system, count DESC`,
}

// MigratePostgres 执行 PostgreSQL 数据库迁移
func MigratePostgres(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// 先确保迁移记录表存在
	if _, err := pool.Exec(ctx, postgresMigrations[0]); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 获取当前版本
	var currentVersion int
	row := pool.QueryRow(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// 执行未应用的迁移（从第2个开始，第1个已执行）
	for i := currentVersion; i < len(postgresMigrations); i++ {
		// 跳过已执行的第一个迁移
		if i == 0 {
			// 记录第一个迁移
			if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT DO NOTHING"); err != nil {
				return fmt.Errorf("failed to record migration 1: %w", err)
			}
			continue
		}

		if _, err := pool.Exec(ctx, postgresMigrations[i]); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}

		// 记录迁移
		if _, err := pool.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", i+1); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", i+1, err)
		}
	}

	return nil
}
