PostgreSQL 架构设计
部署架构
┌─────────────────────────────────────────────────────────────────┐
│                        本地开发环境                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              Docker Compose                              │   │
│  │  ┌─────────────────┐    ┌─────────────────────────┐     │   │
│  │  │   PostgreSQL    │    │     Insight Hub API     │     │   │
│  │  │   Port: 5432    │◄───│     Port: 8080          │     │   │
│  │  │                 │    │                         │     │   │
│  │  │  Database:      │    │  - REST API             │     │   │
│  │  │  insight_hub    │    │  - Dedup Service        │     │   │
│  │  │                 │    │  - Query Engine         │     │   │
│  │  │  User: insight  │    │                         │     │   │
│  │  │                 │    │  Go + pgx driver        │     │   │
│  │  │  Volume:        │    │                         │     │   │
│  │  │  pg_data/       │    └─────────────────────────┘     │   │
│  │  └─────────────────┘                                     │   │
│  │                                                          │   │
│  │  ┌─────────────────────────────────────────────────┐    │   │
│  │  │  pgAdmin (Optional)                              │    │   │
│  │  │  Port: 5050                                      │    │   │
│  │  │  Web UI for database management                  │    │   │
│  │  └─────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ~/.insight-hub/                                               │
│  ├── config.yaml                                               │
│  ├── docker-compose.yaml                                       │
│  └── pg_data/  (PostgreSQL data volume)                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘


数据模型设计
- **v3 (最新):** [data-model-v3.md](./data-model-v3.md) — 个人数据中枢设计（2026-03-01）
- **v2 (已实现):** [data-model-v2.md](./data-model-v2.md) — Agent 专用设计（已废弃）

> ⚠️ **重要决策 (2026-03-01):** 项目重新定位为"个人数据中枢"，统一存储所有 AI 工具产生的数据。
> - 共享部署（一个 Insight Hub 服务所有 AI 工具）
> - 个人使用（无需多租户）
> - 开放类型（不限制数据类型）
> - source_system 区分来源（openclaw, picoclaw, opencode, claude-code, codex...）

---

数据库 Schema 设计
-- ============================================
-- Insight Hub PostgreSQL Schema
-- Version: 2.0.0
-- 基于 data-model-v2.md
-- ============================================

-- 启用扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";     -- 模糊搜索
-- CREATE EXTENSION IF NOT EXISTS "vector";   -- 向量搜索（未来）

-- ============================================
-- 核心表：Items
-- 
-- type 枚举说明：
--   - brief: 简报（每日简报、突发简报等）
--   - news: 新闻（抓取的新闻文章）
--   - tracking: 追踪（话题追踪、趋势追踪）
--   - insight: 洞察（观察、预测、问题）
--   - log: 日志（Agent 操作日志）
--
-- source_metadata 字段规范：
--   详见 docs/data-model-v2.md
--   - news: author, published_at, category, keywords, summary, sentiment, importance
--   - brief: brief_type, period_start, period_end, sent_to, sent_at, news_count, topics
--   - tracking: track_id, track_type, topic, status, last_check, check_interval, history[]
--   - insight: insight_type, confidence, related_items, actionable, action_taken
--   - log: log_type, action, result, duration_ms, error
--
-- 扩展路径：
--   - 阶段 1 (MVP): 单表 + JSONB
--   - 阶段 2: 性能优化（提取高频字段）
--   - 阶段 3: 数据拆分（按 type 拆表）
-- ============================================
CREATE TABLE items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(32) NOT NULL,           -- brief, news, tracking, insight, log
    agent           VARCHAR(32) NOT NULL,           -- news, tech-doc, ...
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    source          VARCHAR(64),                    -- 来源标识
    source_url      TEXT,                           -- 原始 URL
    source_metadata JSONB DEFAULT '{}',             -- 灵活元数据
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 约束
    CONSTRAINT items_type_check CHECK (type IN ('brief', 'news', 'tracking', 'insight', 'log'))
);

-- 索引
CREATE INDEX idx_items_type ON items(type);
CREATE INDEX idx_items_agent ON items(agent);
CREATE INDEX idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX idx_items_created_at ON items(created_at DESC);
CREATE INDEX idx_items_source ON items(source);
CREATE INDEX idx_items_metadata ON items USING GIN (source_metadata);

-- 全文搜索索引
CREATE INDEX idx_items_fulltext ON items 
USING GIN (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, '')));

-- 模糊搜索索引（用于标题相似度）
CREATE INDEX idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);

-- ============================================
-- 标签表：Tags
-- ============================================
CREATE TABLE tags (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tags_name ON tags(name);

-- ============================================
-- Item-Tag 关联表
-- ============================================
CREATE TABLE item_tags (
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    tag_id      INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    PRIMARY KEY (item_id, tag_id)
);

-- ============================================
-- 去重索引表：Dedup
-- ============================================
CREATE TABLE dedup (
    id          SERIAL PRIMARY KEY,
    url_hash    VARCHAR(128) NOT NULL UNIQUE,   -- SHA256
    url         TEXT NOT NULL,
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE SET NULL,
    agent       VARCHAR(32) NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at  TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_dedup_url_hash ON dedup(url_hash);
CREATE INDEX idx_dedup_agent ON dedup(agent);
CREATE INDEX idx_dedup_expires_at ON dedup(expires_at);

-- ============================================
-- 触发器：自动更新 updated_at
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER items_updated_at
    BEFORE UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ============================================
-- 清理过期去重记录（定时任务调用）
-- ============================================
CREATE OR REPLACE FUNCTION cleanup_expired_dedup()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM dedup WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 视图：活跃追踪对象
-- ============================================
CREATE VIEW active_trackings AS
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

-- ============================================
-- 视图：最近简报
-- ============================================
CREATE VIEW recent_briefs AS
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
-- 初始化标签
-- ============================================
INSERT INTO tags (name) VALUES
    ('agent:news'),
    ('agent:tech-doc'),
    ('每日简报'),
    ('追踪话题'),
    ('问题追踪'),
    ('趋势追踪'),
    ('洞察笔记')
ON CONFLICT (name) DO NOTHING;


Docker Compose 配置
# ~/.insight-hub/docker-compose.yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: insight-hub-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: insight_hub
      POSTGRES_USER: insight
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-insight_local_dev}
      PGDATA: /var/lib/postgresql/data/pgdata
    volumes:
      - ./pg_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql:ro
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U insight -d insight_hub"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Insight Hub API（未来部署）
  # api:
  #   build: ./api
  #   container_name: insight-hub-api
  #   restart: unless-stopped
  #   environment:
  #     DATABASE_URL: postgres://insight:${POSTGRES_PASSWORD:-insight_local_dev}@postgres:5432/insight_hub?sslmode=disable
  #   ports:
  #     - "8080:8080"
  #   depends_on:
  #     postgres:
  #       condition: service_healthy

  # pgAdmin（可选，数据库管理 UI）
  pgadmin:
    image: dpage/pgadmin4:latest
    container_name: insight-hub-pgadmin
    restart: unless-stopped
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@insight-hub.local
      PGADMIN_DEFAULT_PASSWORD: ${PGADMIN_PASSWORD:-admin}
      PGADMIN_LISTEN_PORT: 5050
    ports:
      - "5050:5050"
    volumes:
      - ./pgadmin_data:/var/lib/pgadmin
    depends_on:
      - postgres
    profiles:
      - admin  # docker compose --profile admin up

networks:
  default:
    name: insight-hub-network


应用配置
# ~/.insight-hub/config.yaml
database:
  host: localhost
  port: 5432
  name: insight_hub
  user: insight
  password: ${POSTGRES_PASSWORD:-insight_local_dev}
  sslmode: disable
  # 连接池
  max_conns: 10
  min_conns: 2
  max_conn_lifetime: 1h
  max_conn_idle_time: 10m

server:
  host: localhost
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

agents:
  - name: news
    enabled: true
    dedup_ttl: 168h  # 7天
  - name: tech-doc
    enabled: true
    dedup_ttl: 168h

logging:
  level: info
  format: json
  output: stdout

# 全文搜索配置
search:
  language: english
  limit: 100

# 定时任务
scheduler:
  dedup_cleanup: "0 3 * * *"  # 每天凌晨3点清理过期去重记录


Go 数据访问层设计
// internal/store/store.go
package store

import (
    "context"
    "time"
    
    "github.com/jackc/pgx/v5/pgxpool"
)

// ============================================
// 类型常量
// ============================================

const (
    ItemTypeBrief    = "brief"     // 简报
    ItemTypeNews     = "news"      // 新闻
    ItemTypeTracking = "tracking"  // 追踪
    ItemTypeInsight  = "insight"   // 洞察
    ItemTypeLog      = "log"       // 日志
)

// ============================================
// Agent 常量
// ============================================

const (
    AgentNews    = "news-agent"
    AgentTechDoc = "tech-doc-agent"
)

// ============================================
// Store 接口
// ============================================

type Store struct {
    pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*Store, error) {
    pool, err := pgxpool.New(ctx, connString)
    if err != nil {
        return nil, err
    }
    return &Store{pool: pool}, nil
}

func (s *Store) Close() {
    s.pool.Close()
}

// ============================================
// Item 结构体
// ============================================

type Item struct {
    ID             string                 `json:"id"`
    Type           string                 `json:"type"`           // brief, news, tracking, insight, log
    Agent          string                 `json:"agent"`          // news-agent, tech-doc-agent
    Title          string                 `json:"title"`
    Content        string                 `json:"content"`
    Source         string                 `json:"source"`
    SourceURL      string                 `json:"source_url"`
    SourceMetadata map[string]interface{} `json:"source_metadata"`
    OccurredAt     *time.Time             `json:"occurred_at"`
    CreatedAt      time.Time              `json:"created_at"`
    UpdatedAt      time.Time              `json:"updated_at"`
    Tags           []string               `json:"tags"`
}

// ============================================
// source_metadata 辅助结构体
// ============================================

// NewsMetadata 新闻元数据
type NewsMetadata struct {
    Author      string   `json:"author"`
    PublishedAt string   `json:"published_at"`
    Category    string   `json:"category"`
    Keywords    []string `json:"keywords"`
    Summary     string   `json:"summary"`
    Sentiment   string   `json:"sentiment"`
    Importance  float64  `json:"importance"`
}

// BriefMetadata 简报元数据
type BriefMetadata struct {
    BriefType   string   `json:"brief_type"`   // daily, weekly, breaking
    PeriodStart string   `json:"period_start"`
    PeriodEnd   string   `json:"period_end"`
    SentTo      []string `json:"sent_to"`
    SentAt      string   `json:"sent_at"`
    NewsCount   int      `json:"news_count"`
    Topics      []string `json:"topics"`
}

// TrackingMetadata 追踪元数据
type TrackingMetadata struct {
    TrackID       string                `json:"track_id"`
    TrackType     string                `json:"track_type"`     // trend, issue, topic
    Topic         string                `json:"topic"`
    Status        string                `json:"status"`         // active, paused, completed
    LastCheck     string                `json:"last_check"`
    CheckInterval string                `json:"check_interval"`
    History       []TrackingHistoryItem `json:"history"`        // 最多 30 条
}

// TrackingHistoryItem 追踪历史项
type TrackingHistoryItem struct {
    Date         string  `json:"date"`
    Summary      string  `json:"summary"`
    Importance   float64 `json:"importance"`
    SourceCount  int     `json:"source_count"`
}

// InsightMetadata 洞察元数据
type InsightMetadata struct {
    InsightType  string   `json:"insight_type"`  // observation, prediction, question
    Confidence   float64  `json:"confidence"`
    RelatedItems []string `json:"related_items"`
    Actionable   bool     `json:"actionable"`
    ActionTaken  string   `json:"action_taken"`
}

// LogMetadata 日志元数据
type LogMetadata struct {
    LogType    string `json:"log_type"`    // agent_action, system, error
    Action     string `json:"action"`
    Result     string `json:"result"`
    DurationMs int    `json:"duration_ms"`
    Error      string `json:"error"`
}

// ============================================
// Item 操作
// ============================================

type Item struct {
    ID             string                 `json:"id"`
    Type           string                 `json:"type"`
    Agent          string                 `json:"agent"`
    Title          string                 `json:"title"`
    Content        string                 `json:"content"`
    Source         string                 `json:"source"`
    SourceURL      string                 `json:"source_url"`
    SourceMetadata map[string]interface{} `json:"source_metadata"`
    OccurredAt     *time.Time             `json:"occurred_at"`
    CreatedAt      time.Time              `json:"created_at"`
    UpdatedAt      time.Time              `json:"updated_at"`
    Tags           []string               `json:"tags"`
}

type QueryFilter struct {
    Type       string
    Agent      string
    TrackType  string  // for source_metadata->>'track_type'
    Status     string  // for source_metadata->>'status'
    DateFrom   *time.Time
    DateTo     *time.Time
    Search     string  // 全文搜索
    Limit      int
    Offset     int
}

func (s *Store) CreateItem(ctx context.Context, item *Item) error {
    query := `
        INSERT INTO items (id, type, agent, title, content, source, source_url, source_metadata, occurred_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    _, err := s.pool.Exec(ctx, query,
        item.ID, item.Type, item.Agent, item.Title, item.Content,
        item.Source, item.SourceURL, item.SourceMetadata, item.OccurredAt,
    )
    return err
}

// UpdateTrackingHistory 更新追踪话题的历史记录
// 限制 history 数组最多 30 条
func (s *Store) UpdateTrackingHistory(ctx context.Context, itemID string, historyItem TrackingHistoryItem) error {
    query := `
        UPDATE items
        SET source_metadata = jsonb_set(
            source_metadata,
            '{history}',
            (
                SELECT jsonb_agg(elem)
                FROM (
                    SELECT elem
                    FROM jsonb_array_elements(source_metadata->'history') elem
                    UNION ALL
                    SELECT $2::jsonb
                    ORDER BY (elem->>'date') DESC
                    LIMIT 30
                ) sub
            )
        )
        WHERE id = $1
    `
    historyJSON, _ := json.Marshal(historyItem)
    _, err := s.pool.Exec(ctx, query, itemID, historyJSON)
    return err
}

func (s *Store) GetItem(ctx context.Context, id string) (*Item, error) {
    query := `
        SELECT i.id, i.type, i.agent, i.title, i.content, i.source, i.source_url,
               i.source_metadata, i.occurred_at, i.created_at, i.updated_at,
               COALESCE(array_agg(t.name) FILTER (WHERE t.name IS NOT NULL), '{}') as tags
        FROM items i
        LEFT JOIN item_tags it ON i.id = it.item_id
        LEFT JOIN tags t ON it.tag_id = t.id
        WHERE i.id = $1
        GROUP BY i.id
    `
    // ... 实现查询逻辑
}

func (s *Store) QueryItems(ctx context.Context, filter QueryFilter) ([]*Item, error) {
    // 动态构建查询
    // 支持按 type, agent, source_metadata 字段筛选
    // 支持全文搜索
    // ...
}

func (s *Store) UpdateItem(ctx context.Context, id string, updates map[string]interface{}) error {
    // 动态构建 UPDATE 语句
    // ...
}

func (s *Store) DeleteItem(ctx context.Context, id string) error {
    _, err := s.pool.Exec(ctx, "DELETE FROM items WHERE id = $1", id)
    return err
}

// ============================================
// 标签操作
// ============================================

func (s *Store) AddTags(ctx context.Context, itemID string, tags []string) error {
    // 批量插入标签关联
    // ...
}

func (s *Store) RemoveTags(ctx context.Context, itemID string, tags []string) error {
    // ...
}

// ============================================
// 去重操作
// ============================================

type DedupRecord struct {
    URLHash   string
    URL       string
    ItemID    string
    Agent     string
    ExpiresAt time.Time
}

func (s *Store) CheckDedup(ctx context.Context, urlHash string) (*DedupRecord, error) {
    query := `
        SELECT url_hash, url, item_id, agent, expires_at
        FROM dedup
        WHERE url_hash = $1 AND expires_at > NOW()
    `
    // ...
}

func (s *Store) RecordDedup(ctx context.Context, record *DedupRecord) error {
    query := `
        INSERT INTO dedup (url_hash, url, item_id, agent, expires_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (url_hash) DO UPDATE SET
            item_id = EXCLUDED.item_id,
            expires_at = EXCLUDED.expires_at
    `
    // ...
}

func (s *Store) CleanupExpiredDedup(ctx context.Context) (int, error) {
    var count int
    err := s.pool.QueryRow(ctx, "SELECT cleanup_expired_dedup()").Scan(&count)
    return count, err
}
API 设计
// internal/api/handlers.go

// POST /api/items
func (h *Handler) CreateItem(c *gin.Context) {
    var req struct {
        ID             string                 `json:"id"`
        Type           string                 `json:"type" binding:"required"`
        Agent          string                 `json:"agent" binding:"required"`
        Title          string                 `json:"title" binding:"required"`
        Content        string                 `json:"content"`
        Source         string                 `json:"source"`
        SourceURL      string                 `json:"source_url"`
        SourceMetadata map[string]interface{} `json:"source_metadata"`
        OccurredAt     *time.Time             `json:"occurred_at"`
        Tags           []string               `json:"tags"`
    }
    // ...
}

// GET /api/items
// Query params: type, agent, track_type, status, date_from, date_to, search, limit, offset
func (h *Handler) QueryItems(c *gin.Context) {
    filter := store.QueryFilter{
        Type:      c.Query("type"),
        Agent:     c.Query("agent"),
        TrackType: c.Query("track_type"),
        Status:    c.Query("status"),
        Search:    c.Query("search"),
        Limit:     100,
    }
    // ...
}

// GET /api/items/:id
func (h *Handler) GetItem(c *gin.Context) {
    // ...
}

// PUT /api/items/:id
func (h *Handler) UpdateItem(c *gin.Context) {
    // ...
}

// DELETE /api/items/:id
func (h *Handler) DeleteItem(c *gin.Context) {
    // ...
}

// POST /api/dedup/check
func (h *Handler) CheckDedup(c *gin.Context) {
    var req struct {
        URLHash string `json:"url_hash" binding:"required"`
    }
    // ...
}

// POST /api/dedup
func (h *Handler) RecordDedup(c *gin.Context) {
    var req struct {
        URLHash string    `json:"url_hash" binding:"required"`
        URL     string    `json:"url" binding:"required"`
        ItemID  string    `json:"item_id"`
        Agent   string    `json:"agent" binding:"required"`
        TTL     duration  `json:"ttl"`  // 默认 168h
    }
    // ...
}

// GET /api/trackings/active
func (h *Handler) GetActiveTrackings(c *gin.Context) {
    agent := c.Query("agent")
    // 使用 active_trackings 视图
    // ...
}

// GET /api/briefs/recent
func (h *Handler) GetRecentBriefs(c *gin.Context) {
    agent := c.Query("agent")
    // 使用 recent_briefs 视图
    // ...
}


项目结构
insight-hub/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   └── routes.go
│   ├── store/
│   │   ├── store.go
│   │   ├── items.go
│   │   ├── tags.go
│   │   └── dedup.go
│   ├── config/
│   │   └── config.go
│   └── scheduler/
│       └── scheduler.go
├── migrations/
│   └── 001_init.sql
├── scripts/
│   ├── init-db.sh
│   └── backup.sh
├── docker-compose.yaml
├── Dockerfile
├── Makefile
├── config.yaml
└── README.md


---

实施计划
Phase 1: 基础设施搭建（1天）
创建项目结构
编写 Docker Compose + 初始化脚本
启动 PostgreSQL 并验证
配置 pgAdmin（可选）

Phase 2: 数据访问层实现（1天）
实现 Store 接口
编写单元测试
验证 JSONB 查询性能

Phase 3: API 实现（1天）
实现 REST API
编写 API 测试
添加认证（如需要）

Phase 4: Agent 集成（2天）
News Agent 集成
Tech-Doc Agent 集成
端到端测试

---
预估工作量: 5 天


---

## UI v2 实现记录 (2026-03-01)

### 技术选型
- **前端框架**: Vanilla JS (无框架依赖)
- **CSS 框架**: Tailwind CSS CDN
- **字体**: Fira Sans + Fira Code
- **风格**: Dark Mode + Glassmorphism + Bento Grid

### 文件位置
- `web/static/index-v2.html` — 完整单页应用

### 访问方式
- http://localhost:8090/index-v2.html
- http://localhost:8090/v2 (重定向)

### 已实现功能
- ✅ 统计数据展示 (Bento Grid)
- ✅ 类型/来源/状态筛选
- ✅ 全文搜索
- ✅ 排序 (创建时间/更新时间/标题)
- ✅ 分页加载
- ✅ 详情模态框
- ✅ 删除功能
- ✅ 响应式布局
- ✅ 类型感知卡片 (tracking/news 展示不同字段)

### 待实现功能
- [ ] 新建/编辑功能
- [ ] 标签云动态加载
- [ ] 时间线视图 (tracking 类型)
- [ ] 批量操作


---

## 功能路线图 (2026-03-02)

> 详细规划见 [roadmap.md](./roadmap.md)

### 版本规划

#### v1.1 (近期)

| 功能 | 优先级 | 状态 | 设计文档 |
|------|--------|------|---------|
| Markdown 渲染 | P0 | 设计中 | [markdown-rendering-design.md](./markdown-rendering-design.md) |
| 任务看板视图 | P1 | 规划中 | - |
| 基础图表 (Chart.js) | P1 | 规划中 | - |

#### v1.2 (中期)

- 高级搜索
- 时间范围筛选
- Dashboard 页面

#### v2.0 (远期)

- AI 辅助分析
- 知识图谱
- 报告生成

---

## 文档索引

| 文档 | 说明 |
|------|------|
| [roadmap.md](./roadmap.md) | 功能路线图与方案讨论 |
| [markdown-rendering-design.md](./markdown-rendering-design.md) | Markdown 渲染设计方案 |
| [data-model-v3.md](./data-model-v3.md) | 数据模型 v3 |
| [api-guide.md](./api-guide.md) | API 使用指南 |
| [ui-design-v2.md](./ui-design-v2.md) | UI v2 设计文档 |
