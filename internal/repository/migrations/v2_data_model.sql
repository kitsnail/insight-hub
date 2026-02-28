-- ============================================
-- Insight Hub Schema Migration v2
-- 数据模型重构：支持多 Agent + 结构化元数据
-- Date: 2026-03-01
-- ============================================

-- ============================================
-- 1. 添加 agent 字段
-- ============================================
ALTER TABLE items ADD COLUMN IF NOT EXISTS agent VARCHAR(32) NOT NULL DEFAULT 'unknown';

-- 创建 agent 索引
CREATE INDEX IF NOT EXISTS idx_items_agent ON items(agent);

-- ============================================
-- 2. 更新 type 约束
-- ============================================
-- 先删除旧约束
ALTER TABLE items DROP CONSTRAINT IF EXISTS items_type_check;

-- 添加新约束
ALTER TABLE items ADD CONSTRAINT items_type_check 
  CHECK (type IN ('brief', 'news', 'tracking', 'insight', 'log'));

-- ============================================
-- 3. 更新现有数据的 type
-- ============================================
-- 将旧类型映射到新类型
UPDATE items SET type = 'news' WHERE type = 'bookmark';
UPDATE items SET type = 'insight' WHERE type = 'note';
UPDATE items SET type = 'log' WHERE type = 'chat';
UPDATE items SET type = 'log' WHERE type = 'code';
UPDATE items SET type = 'log' WHERE type = 'task';

-- ============================================
-- 4. 更新现有数据的 agent
-- ============================================
-- 根据来源推断 agent
UPDATE items SET agent = 'news-agent' WHERE source LIKE '%news%' OR source LIKE '%rss%' OR source LIKE '%feed%';
UPDATE items SET agent = 'tech-doc-agent' WHERE source LIKE '%github%' OR source LIKE '%doc%' OR source LIKE '%tech%';
UPDATE items SET agent = 'unknown' WHERE agent = 'unknown'; -- 保持默认

-- ============================================
-- 5. 创建新索引
-- ============================================
-- JSONB GIN 索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_items_metadata ON items USING GIN (source_metadata);

-- 全文搜索索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_items_fulltext ON items 
USING GIN (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, '')));

-- ============================================
-- 6. 创建视图
-- ============================================
-- 活跃追踪对象视图
CREATE OR REPLACE VIEW active_trackings AS
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
ORDER BY updated_at DESC;

-- 最近简报视图
CREATE OR REPLACE VIEW recent_briefs AS
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
LIMIT 30;

-- ============================================
-- 7. 更新去重表
-- ============================================
-- 添加 agent 字段到去重表
ALTER TABLE dedup ADD COLUMN IF NOT EXISTS agent VARCHAR(32);
CREATE INDEX IF NOT EXISTS idx_dedup_agent ON dedup(agent);

-- ============================================
-- 8. 清理旧表（可选，保留用于回滚）
-- ============================================
-- 暂不删除，等验证无误后再清理
-- DROP TABLE IF EXISTS insights;
-- DROP TABLE IF EXISTS sources;
-- DROP TABLE IF EXISTS processing_queue;

-- ============================================
-- 验证迁移
-- ============================================
-- 检查 type 分布
SELECT type, COUNT(*) FROM items GROUP BY type;

-- 检查 agent 分布
SELECT agent, COUNT(*) FROM items GROUP BY agent;

-- 检查约束
SELECT conname, pg_get_constraintdef(oid) 
FROM pg_constraint 
WHERE conrelid = 'items'::regclass;
