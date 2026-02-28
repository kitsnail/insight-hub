#!/bin/bash
# 在 postgres-local 中创建 insight_hub 数据库

docker exec -i postgres-local psql -U postgres <<EOF
-- 创建数据库
CREATE DATABASE insight_hub;

-- 创建用户（如果不存在）
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'insight') THEN
        CREATE USER insight WITH PASSWORD 'insight_local_dev';
    END IF;
END
\$\$;

-- 授权
GRANT ALL PRIVILEGES ON DATABASE insight_hub TO insight;

-- 连接到 insight_hub 并创建扩展
\c insight_hub

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- 授权 schema
GRANT ALL ON SCHEMA public TO insight;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO insight;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO insight;

-- 显示数据库列表
\l
EOF
