-- ============================================
-- Insight Hub Schema v3
-- 从零开始创建 v3 数据库结构
-- ============================================

-- 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ============================================
-- 核心表：items
-- ============================================

CREATE TABLE IF NOT EXISTS items (
    -- 主键
    id              VARCHAR(64) PRIMARY KEY,
    
    -- 数据分类（开放）
    type            VARCHAR(64) NOT NULL,           -- 开放类型，不限制
    category        VARCHAR(64),                    -- 可选二级分类
    
    -- 来源追溯
    source_system   VARCHAR(64) NOT NULL,           -- 数据来源系统
    source_agent    VARCHAR(64),                    -- 来源 Agent（可选）
    source_url      TEXT,                           -- 原始 URL（如有）
    
    -- 核心内容
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    summary         TEXT,                           -- 摘要（可选）
    
    -- 灵活元数据
    metadata        JSONB DEFAULT '{}',             -- 结构化元数据
    
    -- 状态与优先级
    status          VARCHAR(32) DEFAULT 'active',   -- active, archived, deleted
    priority        INTEGER DEFAULT 0,              -- 优先级（可选）
    
    -- 时间戳
    occurred_at     TIMESTAMP WITH TIME ZONE,       -- 事件发生时间
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 全文搜索向量
    search_vector   TSVECTOR
);

-- ============================================
-- 索引
-- ============================================

-- 基础索引
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category);
CREATE INDEX IF NOT EXISTS idx_items_source_system ON items(source_system);
CREATE INDEX IF NOT EXISTS idx_items_status ON items(status);
CREATE INDEX IF NOT EXISTS idx_items_priority ON items(priority DESC);

-- 时间索引
CREATE INDEX IF NOT EXISTS idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_created_at ON items(created_at DESC);

-- 元数据索引（GIN）
CREATE INDEX IF NOT EXISTS idx_items_metadata ON items USING GIN (metadata);

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS idx_items_search_vector ON items USING GIN (search_vector);

-- 模糊搜索索引（标题）
CREATE INDEX IF NOT EXISTS idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);

-- 复合索引（常见查询）
CREATE INDEX IF NOT EXISTS idx_items_type_status ON items(type, status);
CREATE INDEX IF NOT EXISTS idx_items_source_type ON items(source_system, type);
CREATE INDEX IF NOT EXISTS idx_items_type_occurred ON items(type, occurred_at DESC);

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

DROP TRIGGER IF EXISTS items_updated_at ON items;
CREATE TRIGGER items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 触发器：自动更新 search_vector
-- ============================================

CREATE OR REPLACE FUNCTION update_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector = to_tsvector('english',
        coalesce(NEW.title, '') || ' ' ||
        coalesce(NEW.content, '') || ' ' ||
        coalesce(NEW.summary, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS items_search_vector ON items;
CREATE TRIGGER items_search_vector
    BEFORE INSERT OR UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_search_vector();

-- ============================================
-- 标签表
-- ============================================

CREATE TABLE IF NOT EXISTS tags (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,
    category    VARCHAR(64),
    color       VARCHAR(32),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category);

-- ============================================
-- Item-Tag 关联表
-- ============================================

CREATE TABLE IF NOT EXISTS item_tags (
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    tag_id      INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (item_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_item_tags_item ON item_tags(item_id);
CREATE INDEX IF NOT EXISTS idx_item_tags_tag ON item_tags(tag_id);

-- ============================================
-- 去重表
-- ============================================

CREATE TABLE IF NOT EXISTS dedup (
    fingerprint VARCHAR(64) PRIMARY KEY,
    item_id     VARCHAR(64) NOT NULL,
    agent       VARCHAR(64),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at  TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_dedup_expires ON dedup(expires_at) WHERE expires_at IS NOT NULL;

-- ============================================
-- 来源系统注册表（可选）
-- ============================================

CREATE TABLE IF NOT EXISTS source_systems (
    name        VARCHAR(64) PRIMARY KEY,
    description TEXT,
    config      JSONB DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 初始化来源系统
INSERT INTO source_systems (name, description) VALUES
    ('openclaw', 'openclaw AI system'),
    ('picoclaw', 'picoclaw AI system'),
    ('opencode', 'opencode development tool'),
    ('claude-code', 'Claude Code assistant'),
    ('codex', 'OpenAI Codex'),
    ('manual', 'Manual entry'),
    ('web', 'Web scraper')
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 视图：活跃任务
-- ============================================

CREATE OR REPLACE VIEW active_tasks AS
SELECT 
    id, title, content, source_system,
    metadata->>'due' AS due,
    metadata->>'status' AS task_status,
    priority, occurred_at, created_at
FROM items
WHERE type = 'task'
  AND status = 'active'
  AND (metadata->>'status' IN ('pending', 'in_progress') OR metadata->>'status' IS NULL)
ORDER BY priority DESC, occurred_at ASC;

-- ============================================
-- 视图：最近决策
-- ============================================

CREATE OR REPLACE VIEW recent_decisions AS
SELECT 
    id, title, source_system,
    metadata->>'outcome' AS outcome,
    metadata->>'rationale' AS rationale,
    occurred_at, created_at
FROM items
WHERE type = 'decision'
  AND status = 'active'
ORDER BY occurred_at DESC
LIMIT 50;

-- ============================================
-- 视图：按来源统计
-- ============================================

CREATE OR REPLACE VIEW stats_by_source AS
SELECT 
    source_system,
    type,
    COUNT(*) AS count,
    MAX(created_at) AS last_created
FROM items
WHERE status = 'active'
GROUP BY source_system, type
ORDER BY source_system, count DESC;
