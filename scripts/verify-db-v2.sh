#!/bin/bash
# 快速验证数据库状态

echo "=== 当前数据库状态 ==="
echo ""

echo "1. 表结构:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "\d items"
echo ""

echo "2. type 分布:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT type, COUNT(*) FROM items GROUP BY type ORDER BY COUNT(*) DESC;"
echo ""

echo "3. 迁移版本:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT * FROM schema_migrations ORDER BY version DESC LIMIT 5;"
echo ""

echo "4. 索引:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "\di items*"
echo ""

echo "5. 视图:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "\dv"
echo ""
