#!/bin/bash
# Insight Hub 数据库迁移脚本 v2
# 执行数据模型重构迁移

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 数据库连接信息
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5433}"
DB_NAME="${DB_NAME:-insight_hub}"
DB_USER="${DB_USER:-insight}"
DB_PASSWORD="${DB_PASSWORD:-insight_local_dev}"

# 迁移文件路径
MIGRATION_FILE="internal/repository/migrations/v2_data_model.sql"

echo -e "${YELLOW}=== Insight Hub 数据库迁移 v2 ===${NC}"
echo ""
echo "数据库: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "用户: ${DB_USER}"
echo ""

# 检查 Docker 容器是否运行
if ! docker ps | grep -q "insight-hub-postgres"; then
    echo -e "${RED}错误: PostgreSQL 容器未运行${NC}"
    echo "请先启动容器: docker-compose up -d postgres"
    exit 1
fi

echo -e "${YELLOW}步骤 1: 备份当前数据...${NC}"
BACKUP_FILE="/tmp/insight_hub_backup_$(date +%Y%m%d_%H%M%S).sql"
docker exec insight-hub-postgres pg_dump -U insight insight_hub > "$BACKUP_FILE"
echo -e "${GREEN}✓ 备份已保存到: $BACKUP_FILE${NC}"
echo ""

echo -e "${YELLOW}步骤 2: 检查当前数据...${NC}"
echo "当前 type 分布:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT type, COUNT(*) FROM items GROUP BY type ORDER BY COUNT(*) DESC;"
echo ""

echo -e "${YELLOW}步骤 3: 执行迁移...${NC}"
echo "正在执行迁移脚本..."
cat "$MIGRATION_FILE" | docker exec -i insight-hub-postgres psql -U insight -d insight_hub
echo ""

echo -e "${YELLOW}步骤 4: 验证迁移结果...${NC}"
echo "新的 type 分布:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT type, COUNT(*) FROM items GROUP BY type ORDER BY COUNT(*) DESC;"
echo ""

echo "agent 分布:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT agent, COUNT(*) FROM items GROUP BY agent ORDER BY COUNT(*) DESC;"
echo ""

echo "约束检查:"
docker exec insight-hub-postgres psql -U insight -d insight_hub -c "SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid = 'items'::regclass;"
echo ""

echo -e "${GREEN}=== 迁移完成 ===${NC}"
echo ""
echo "如果需要回滚，请执行:"
echo "  cat $BACKUP_FILE | docker exec -i insight-hub-postgres psql -U insight -d insight_hub"
echo ""
