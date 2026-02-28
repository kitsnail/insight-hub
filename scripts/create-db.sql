CREATE DATABASE insight_hub;

DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'insight') THEN
        CREATE USER insight WITH PASSWORD 'insight_local_dev';
    END IF;
END
$$;

GRANT ALL PRIVILEGES ON DATABASE insight_hub TO insight;
