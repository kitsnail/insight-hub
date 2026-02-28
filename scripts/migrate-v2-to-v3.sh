#!/bin/bash

# v2 -> v3 迁移脚本

echo "执行 v2 -> v3 迁移..."

docker exec -i insight-hub-postgres psql -U insight -d insight_hub <<'EOF'
-- 1. 添加新字段
ALTER TABLE items ADD COLUMN IF NOT EXISTS category VARCHAR(64);
ALTER TABLE items ADD COLUMN IF NOT EXISTS source_system VARCHAR(64);
ALTER TABLE items ADD COLUMN IF NOT EXISTS source_agent VARCHAR(64);
ALTER TABLE items ADD COLUMN IF NOT EXISTS metadata JSONB;
ALTER TABLE items ADD COLUMN IF NOT EXISTS priority INTEGER DEFAULT 0;
ALTER TABLE items ADD COLUMN IF NOT EXISTS search_vector TSVECTOR;

-- 2. 迁移数据
UPDATE items SET 
    source_system = COALESCE(agent, 'unknown'),
    source_agent = agent,
    metadata = COALESCE(source_metadata, '{}'::jsonb)
WHERE source_system IS NULL;

-- 3. 删除旧的 type 约束
ALTER TABLE items DROP CONSTRAINT IF EXISTS items_type_check;

-- 4. 修改 type 字段长度
ALTER TABLE items ALTER COLUMN type TYPE VARCHAR(64);

-- 5. 创建新索引
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category);
CREATE INDEX IF NOT EXISTS idx_items_source_system ON items(source_system);
CREATE INDEX IF NOT EXISTS idx_items_priority ON items(priority DESC);
CREATE INDEX IF NOT EXISTS idx_items_metadata_v3 ON items USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_items_search_vector_v3 ON items USING GIN (search_vector);

-- 6. 更新搜索向量
UPDATE items SET search_vector = to_tsvector('english',
    coalesce(title, '') || ' ' ||
    coalesce(content, '') || ' ' ||
    coalesce(summary, '')
);

-- 7. 创建来源系统注册表
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

-- 8. 创建视图
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
