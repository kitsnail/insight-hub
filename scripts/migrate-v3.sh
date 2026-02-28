#!/bin/bash

# Insight Hub v3 数据库迁移脚本
# 使用方法: ./scripts/migrate-v3.sh

set -e

echo "Insight Hub v3 数据库迁移"
echo "========================="

# 检查 Docker 容器是否运行
if ! docker ps | grep -q insight-hub-postgres; then
    echo "错误: PostgreSQL 容器未运行"
    echo "请先启动容器: docker start insight-hub-postgres"
    exit 1
fi

# 执行迁移
echo "执行 v3 schema 迁移..."
docker exec -i insight-hub-postgres psql -U insight -d insight_hub <<'EOF'
-- ============================================
-- Insight Hub Schema v3
-- ============================================

-- 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 核心表
CREATE TABLE IF NOT EXISTS items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(64) NOT NULL,
    category        VARCHAR(64),
    source_system   VARCHAR(64) NOT NULL,
    source_agent    VARCHAR(64),
    source_url      TEXT,
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    summary         TEXT,
    metadata        JSONB DEFAULT '{}',
    status          VARCHAR(32) DEFAULT 'active',
    priority        INTEGER DEFAULT 0,
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    search_vector   TSVECTOR
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_items_type ON items(type);
CREATE INDEX IF NOT EXISTS idx_items_source_system ON items(source_system);
CREATE INDEX IF NOT EXISTS idx_items_status ON items(status);
CREATE INDEX IF NOT EXISTS idx_items_created_at ON items(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_items_metadata ON items USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_items_search_vector ON items USING GIN (search_vector);
CREATE INDEX IF NOT EXISTS idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);

-- 触发器
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

-- 标签表
CREATE TABLE IF NOT EXISTS tags (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,
    category    VARCHAR(64),
    color       VARCHAR(32),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS item_tags (
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    tag_id      INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (item_id, tag_id)
);

-- 去重表
CREATE TABLE IF NOT EXISTS dedup (
    fingerprint VARCHAR(64) PRIMARY KEY,
    item_id     VARCHAR(64) NOT NULL,
    agent       VARCHAR(64),
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at  TIMESTAMP WITH TIME ZONE
);

-- 来源系统注册表
CREATE TABLE IF NOT EXISTS source_systems (
    name        VARCHAR(64) PRIMARY KEY,
    description TEXT,
    config      JSONB DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO source_systems (name, description) VALUES
    ('openclaw', 'openclaw AI system'),
    ('picoclaw', 'picoclaw AI system'),
    ('opencode', 'opencode development tool'),
    ('claude-code', 'Claude Code assistant'),
    ('codex', 'OpenAI Codex'),
    ('manual', 'Manual entry'),
    ('web', 'Web scraper')
ON CONFLICT (name) DO NOTHING;

-- 视图
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

EOF

echo "✅ 迁移完成"
echo ""
echo "验证表结构:"
docker exec -i insight-hub-postgres psql -U insight -d insight_hub -c "\d items"
