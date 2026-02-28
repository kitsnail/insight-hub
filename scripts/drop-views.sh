#!/bin/bash

# 删除旧视图

docker exec -i insight-hub-postgres psql -U insight -d insight_hub <<'EOF'
DROP VIEW IF EXISTS active_trackings CASCADE;
DROP VIEW IF EXISTS recent_briefs CASCADE;
DROP VIEW IF EXISTS stats_by_source CASCADE;
EOF

echo "✅ 视图已删除"
