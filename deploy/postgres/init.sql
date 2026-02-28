-- ============================================
-- Insight Hub PostgreSQL Schema
-- Version: 1.0.0
-- ============================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================
-- 核心表：Items
-- ============================================
CREATE TABLE items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(32) NOT NULL,
    agent           VARCHAR(32) NOT NULL DEFAULT 'default',
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    summary         TEXT,
    source          VARCHAR(64),
    source_url      TEXT,
    source_metadata JSONB DEFAULT '{}',
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status          VARCHAR(32) DEFAULT 'active',
    
    -- 约束
    CONSTRAINT items_type_check CHECK (type IN ('brief', 'news', 'tracking', 'insight', 'log', 'bookmark', 'note', 'chat', 'code', 'task'))
);

-- 索引
CREATE INDEX idx_items_type ON items(type);
CREATE INDEX idx_items_agent ON items(agent);
CREATE INDEX idx_items_status ON items(status);
CREATE INDEX idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX idx_items_created_at ON items(created_at DESC);
CREATE INDEX idx_items_source ON items(source);
CREATE INDEX idx_items_metadata ON items USING GIN (source_metadata);

-- 全文搜索索引
CREATE INDEX idx_items_fulltext ON items 
USING GIN (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, '') || ' ' || coalesce(summary, '')));

-- 模糊搜索索引
CREATE INDEX idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);

-- ============================================
-- 标签表：Tags
-- ============================================
CREATE TABLE tags (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,
    category    VARCHAR(64),
    color       VARCHAR(16),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tags_name ON tags(name);
CREATE INDEX idx_tags_category ON tags(category);

-- ============================================
-- Item-Tag 关联表
-- ============================================
CREATE TABLE item_tags (
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    tag_id      INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (item_id, tag_id)
);

-- ============================================
-- 去重索引表：Dedup
-- ============================================
CREATE TABLE dedup (
    id          SERIAL PRIMARY KEY,
    url_hash    VARCHAR(128) NOT NULL UNIQUE,
    url         TEXT NOT NULL,
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE SET NULL,
    agent       VARCHAR(32) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_dedup_url_hash ON dedup(url_hash);
CREATE INDEX idx_dedup_agent ON dedup(agent);
CREATE INDEX idx_dedup_expires_at ON dedup(expires_at);

-- ============================================
-- 任务表：Tasks（保留原有功能）
-- ============================================
CREATE TABLE tasks (
    id          VARCHAR(64) PRIMARY KEY,
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE SET NULL,
    title       VARCHAR(512) NOT NULL,
    description TEXT,
    status      VARCHAR(32) DEFAULT 'todo',
    priority    VARCHAR(16) DEFAULT 'medium',
    due_date    TIMESTAMP WITH TIME ZONE,
    started_at  TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    recurrence  VARCHAR(64),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT tasks_status_check CHECK (status IN ('todo', 'in_progress', 'done', 'cancelled')),
    CONSTRAINT tasks_priority_check CHECK (priority IN ('low', 'medium', 'high', 'urgent'))
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);

-- ============================================
-- 洞察表：Insights（保留原有功能）
-- ============================================
CREATE TABLE insights (
    id           VARCHAR(64) PRIMARY KEY,
    type         VARCHAR(32) NOT NULL,
    title        VARCHAR(512) NOT NULL,
    content      TEXT,
    period_start TIMESTAMP WITH TIME ZONE,
    period_end   TIMESTAMP WITH TIME ZONE,
    item_ids     JSONB DEFAULT '[]',
    generated_by VARCHAR(64),
    status       VARCHAR(32) DEFAULT 'active',
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT insights_type_check CHECK (type IN ('weekly_report', 'trend', 'reminder', 'suggestion'))
);

CREATE INDEX idx_insights_type ON insights(type);
CREATE INDEX idx_insights_period ON insights(period_start, period_end);

-- ============================================
-- 数据源表：Sources（保留原有功能）
-- ============================================
CREATE TABLE sources (
    id          VARCHAR(64) PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    type        VARCHAR(32) NOT NULL,
    config      JSONB DEFAULT '{}',
    enabled     BOOLEAN DEFAULT true,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================
-- 处理队列表：ProcessingQueue（保留原有功能）
-- ============================================
CREATE TABLE processing_queue (
    id           SERIAL PRIMARY KEY,
    item_id      VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    action       VARCHAR(32) NOT NULL,
    status       VARCHAR(32) DEFAULT 'pending',
    error        TEXT,
    retry_count  INTEGER DEFAULT 0,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT processing_action_check CHECK (action IN ('summarize', 'tag', 'relate', 'insight')),
    CONSTRAINT processing_status_check CHECK (status IN ('pending', 'processing', 'done', 'failed'))
);

CREATE INDEX idx_processing_queue_status ON processing_queue(status);
CREATE INDEX idx_processing_queue_item ON processing_queue(item_id);

-- ============================================
-- 触发器：自动更新 updated_at
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER sources_updated_at
    BEFORE UPDATE ON sources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 清理过期去重记录
-- ============================================
CREATE OR REPLACE FUNCTION cleanup_expired_dedup()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM dedup WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 视图：活跃追踪对象
-- ============================================
CREATE VIEW active_trackings AS
SELECT 
    id,
    agent,
    title,
    source_metadata->>'track_id' AS track_id,
    source_metadata->>'track_type' AS track_type,
    source_metadata->>'status' AS track_status,
    source_metadata->>'last_check' AS last_check,
    occurred_at,
    updated_at
FROM items
WHERE type = 'tracking'
  AND (source_metadata->>'status' IS NULL OR source_metadata->>'status' = 'active')
ORDER BY updated_at DESC;

-- ============================================
-- 视图：最近简报
-- ============================================
CREATE VIEW recent_briefs AS
SELECT 
    id,
    agent,
    title,
    occurred_at,
    source_metadata->>'sent_to' AS sent_to,
    source_metadata->>'sent_at' AS sent_at,
    created_at
FROM items
WHERE type = 'brief'
ORDER BY occurred_at DESC
LIMIT 100;

-- ============================================
-- 初始化标签
-- ============================================
INSERT INTO tags (name, category) VALUES
    ('agent:news', 'system'),
    ('agent:tech-doc', 'system'),
    ('每日简报', 'category'),
    ('追踪话题', 'category'),
    ('问题追踪', 'category'),
    ('趋势追踪', 'category'),
    ('洞察笔记', 'category')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 授权
-- ============================================
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO insight;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO insight;
