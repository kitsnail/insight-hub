#!/bin/bash

# 删除 type 约束

docker exec -i insight-hub-postgres psql -U insight -d insight_hub -c "ALTER TABLE items DROP CONSTRAINT IF EXISTS items_type_check;"

echo "✅ 约束已删除"
