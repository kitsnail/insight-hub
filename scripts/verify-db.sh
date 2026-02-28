#!/bin/bash
# 验证 PostgreSQL 数据库

echo "=== 验证 PostgreSQL 数据库 ==="
echo ""

echo "--- 表列表 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "\dt"

echo ""
echo "--- 扩展列表 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "\dx"

echo ""
echo "--- 视图列表 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "\dv"

echo ""
echo "--- 标签数据 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT * FROM tags;"

echo ""
echo "--- 测试 JSONB 查询 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "
INSERT INTO items (id, type, agent, title, source_metadata) 
VALUES ('test-001', 'brief', 'news', 'Test Brief', '{\"sent_to\": \"discord#news\"}')
ON CONFLICT (id) DO NOTHING;
"

docker exec insight-hub-postgres psql -U insight -d insight_hub -c "
SELECT id, type, agent, title, source_metadata->>'sent_to' as sent_to 
FROM items 
WHERE id = 'test-001';
"

echo ""
echo "--- 测试全文搜索 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "
SELECT id, title 
FROM items 
WHERE to_tsvector('english', title) @@ to_tsquery('english', 'test');
"

echo ""
echo "--- 清理测试数据 ---"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "DELETE FROM items WHERE id = 'test-001';"

echo ""
echo "=== 验证完成 ==="
