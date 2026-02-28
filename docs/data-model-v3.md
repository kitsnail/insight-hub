# 数据模型 v3 - 个人数据中枢

> **版本:** 3.0.0  
> **日期:** 2026-03-01  
> **状态:** 设计中

---

## 设计目标

### 核心定位

**Insight Hub = 个人数据中枢**

统一存储、管理、查询所有 AI 工具产生的数据，为个人决策和洞察提供数据支撑。

### 使用场景

```
┌─────────────────────────────────────────────────────────────┐
│                      kitsnail 的数字世界                     │
│                                                             │
│                      Insight Hub                            │
│                    (个人数据中枢)                            │
│                                                             │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │ openclaw │ │ picoclaw │ │ opencode │ │ claude   │ ...  │
│  │          │ │          │ │          │ │ code     │      │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘      │
│       │            │            │            │             │
│       └────────────┴────────────┴────────────┘             │
│                         ↓                                   │
│              统一存储、查询、分析                            │
│                         ↓                                   │
│              决策支持、洞察挖掘                              │
└─────────────────────────────────────────────────────────────┘
```

### 设计原则

| 原则 | 说明 |
|------|------|
| **开放性** | 不限制数据类型，任何 AI 工具都能存 |
| **统一性** | 一套 API，所有工具共用 |
| **可追溯** | 记录数据来源，知道谁产生的 |
| **可分析** | 支持统计、趋势、聚合 |
| **可扩展** | 预留向量搜索、知识图谱能力 |

---

## 数据模型

### 核心表：items

```sql
CREATE TABLE items (
    -- 主键
    id              VARCHAR(64) PRIMARY KEY,
    
    -- 数据分类（开放）
    type            VARCHAR(64) NOT NULL,           -- 开放类型，不限制
    category        VARCHAR(64),                    -- 可选二级分类
    
    -- 来源追溯
    source_system   VARCHAR(64) NOT NULL,           -- 数据来源系统
    source_agent    VARCHAR(64),                    -- 来源 Agent（可选）
    source_url      TEXT,                           -- 原始 URL（如有）
    
    -- 核心内容
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    summary         TEXT,                           -- 摘要（可选）
    
    -- 灵活元数据
    metadata        JSONB DEFAULT '{}',             -- 结构化元数据
    
    -- 状态与优先级
    status          VARCHAR(32) DEFAULT 'active',   -- active, archived, deleted
    priority        INTEGER DEFAULT 0,              -- 优先级（可选）
    
    -- 时间戳
    occurred_at     TIMESTAMP WITH TIME ZONE,       -- 事件发生时间
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 全文搜索向量
    search_vector   TSVECTOR
);
```

### 字段说明

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `id` | VARCHAR(64) | 主键，建议用 `{source_system}:{uuid}` | `openclaw:item:abc123` |
| `type` | VARCHAR(64) | 数据类型（开放） | `memory`, `note`, `news`, `decision`, `task` |
| `category` | VARCHAR(64) | 二级分类（可选） | `daily`, `weekly`, `breaking` |
| `source_system` | VARCHAR(64) | 数据来源系统 | `openclaw`, `picoclaw`, `opencode`, `claude-code` |
| `source_agent` | VARCHAR(64) | 来源 Agent（可选） | `news-agent`, `memory-agent` |
| `source_url` | TEXT | 原始 URL | `https://...` |
| `title` | VARCHAR(512) | 标题 | 必填 |
| `content` | TEXT | 内容主体 | 可为空 |
| `summary` | TEXT | 摘要 | 可选 |
| `metadata` | JSONB | 灵活元数据 | 见下文 |
| `status` | VARCHAR(32) | 状态 | `active`, `archived`, `deleted` |
| `priority` | INTEGER | 优先级 | 0=普通, 1=重要, 2=紧急 |
| `occurred_at` | TIMESTAMP | 事件发生时间 | 新闻发布时间、决策时间等 |
| `created_at` | TIMESTAMP | 创建时间 | 自动 |
| `updated_at` | TIMESTAMP | 更新时间 | 自动 |

---

## type 类型设计

### 开放类型

**不限制枚举，任何字符串都可以。** 但建议遵循以下约定：

### 常见类型

| type | 说明 | metadata 示例 |
|------|------|--------------|
| `memory` | 对话记忆 | `{"session_id": "xxx", "turn": 5}` |
| `note` | 笔记 | `{"tags": ["k8s", "devops"]}` |
| `news` | 新闻 | `{"author": "xxx", "published_at": "..."}` |
| `decision` | 决策记录 | `{"context": "...", "alternatives": [...]}` |
| `task` | 任务 | `{"due": "...", "status": "pending"}` |
| `insight` | 洞察 | `{"confidence": 0.8, "related_items": [...]}` |
| `tracking` | 追踪 | `{"topic": "...", "history": [...]}` |
| `brief` | 简报 | `{"period": "daily", "news_count": 10}` |
| `log` | 日志 | `{"action": "...", "result": "success"}` |
| `code` | 代码片段 | `{"language": "go", "file": "main.go"}` |
| `bookmark` | 书签 | `{"url": "...", "tags": [...]}` |

### 类型命名约定

```
- 使用小写
- 用连字符分隔：tech-note, daily-brief
- 或用下划线：tech_note, daily_brief
- 保持一致性（选择一种风格）
```

---

## source_system 设计

### 已知来源系统

| source_system | 说明 | 典型数据类型 |
|---------------|------|-------------|
| `openclaw` | openclaw AI 系统 | memory, decision, note |
| `picoclaw` | picoclaw AI 系统 | memory, note, task |
| `opencode` | opencode 工具 | code, log, decision |
| `claude-code` | Claude Code | memory, code, decision |
| `codex` | Codex | code, log |
| `manual` | 手动录入 | note, task, bookmark |
| `web` | Web 抓取 | news, bookmark |

### 来源系统注册

建议维护一个来源系统注册表（可选）：

```sql
CREATE TABLE source_systems (
    name        VARCHAR(64) PRIMARY KEY,
    description TEXT,
    config      JSONB DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 初始化
INSERT INTO source_systems (name, description) VALUES
    ('openclaw', 'openclaw AI system'),
    ('picoclaw', 'picoclaw AI system'),
    ('opencode', 'opencode development tool'),
    ('claude-code', 'Claude Code assistant'),
    ('codex', 'OpenAI Codex'),
    ('manual', 'Manual entry'),
    ('web', 'Web scraper');
```

---

## metadata 设计

### 设计原则

1. **灵活但结构化** - 用 JSONB，但建议定义 schema
2. **可查询** - 常用字段建索引
3. **可扩展** - 新字段随时加

### 通用 metadata 字段

```json
{
  "tags": ["k8s", "devops"],           // 标签
  "importance": 0.8,                    // 重要性 0-1
  "confidence": 0.9,                    // 置信度 0-1
  "related_items": ["id1", "id2"],      // 关联 item
  "source": {                           // 来源详情
    "author": "xxx",
    "published_at": "2026-03-01T00:00:00Z"
  }
}
```

### 按类型的 metadata 示例

#### memory (对话记忆)
```json
{
  "session_id": "openclaw:session:abc123",
  "turn": 5,
  "role": "assistant",
  "tokens": 256
}
```

#### decision (决策记录)
```json
{
  "context": "选择 PostgreSQL 作为主数据库",
  "alternatives": [
    {"option": "MySQL", "reason": "功能不足"},
    {"option": "MongoDB", "reason": "不适合场景"}
  ],
  "outcome": "PostgreSQL",
  "rationale": "支持 JSONB、全文搜索、扩展性好"
}
```

#### news (新闻)
```json
{
  "author": "CNBC",
  "published_at": "2026-03-01T10:00:00Z",
  "category": "technology",
  "keywords": ["AI", "Kubernetes"],
  "sentiment": "positive",
  "importance": 0.7
}
```

#### task (任务)
```json
{
  "due": "2026-03-05T18:00:00Z",
  "status": "pending",
  "assignee": "kitsnail",
  "parent_task": "openclaw:task:xyz789"
}
```

---

## 索引设计

```sql
-- 基础索引
CREATE INDEX idx_items_type ON items(type);
CREATE INDEX idx_items_category ON items(category);
CREATE INDEX idx_items_source_system ON items(source_system);
CREATE INDEX idx_items_status ON items(status);
CREATE INDEX idx_items_priority ON items(priority DESC);

-- 时间索引
CREATE INDEX idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX idx_items_created_at ON items(created_at DESC);

-- 元数据索引（GIN）
CREATE INDEX idx_items_metadata ON items USING GIN (metadata);

-- 全文搜索索引
CREATE INDEX idx_items_search_vector ON items USING GIN (search_vector);

-- 模糊搜索索引（标题）
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);

-- 复合索引（常见查询）
CREATE INDEX idx_items_type_status ON items(type, status);
CREATE INDEX idx_items_source_type ON items(source_system, type);
CREATE INDEX idx_items_type_occurred ON items(type, occurred_at DESC);
```

---

## 视图设计

### 活跃项目/任务

```sql
CREATE VIEW active_tasks AS
SELECT 
    id, title, content, source_system,
    metadata->>'due' AS due,
    metadata->>'status' AS task_status,
    priority, occurred_at, created_at
FROM items
WHERE type = 'task'
  AND status = 'active'
  AND (metadata->>'status' IN ('pending', 'in_progress') OR metadata->>'status' IS NULL)
ORDER BY priority DESC, occurred_at ASC;
```

### 最近决策

```sql
CREATE VIEW recent_decisions AS
SELECT 
    id, title, source_system,
    metadata->>'outcome' AS outcome,
    metadata->>'rationale' AS rationale,
    occurred_at, created_at
FROM items
WHERE type = 'decision'
  AND status = 'active'
ORDER BY occurred_at DESC
LIMIT 50;
```

### 按来源统计

```sql
CREATE VIEW stats_by_source AS
SELECT 
    source_system,
    type,
    COUNT(*) AS count,
    MAX(created_at) AS last_created
FROM items
WHERE status = 'active'
GROUP BY source_system, type
ORDER BY source_system, count DESC;
```

---

## API 设计

### 统一 CRUD API

```go
// POST /api/v1/items - 创建
// GET  /api/v1/items - 查询
// GET  /api/v1/items/:id - 获取
// PUT  /api/v1/items/:id - 更新
// DELETE /api/v1/items/:id - 删除（软删除）
```

### 查询参数

```
GET /api/v1/items?
    type=memory&                    // 按类型
    source_system=openclaw&         // 按来源
    status=active&                  // 按状态
    date_from=2026-03-01&           // 时间范围
    date_to=2026-03-02&
    search=kubernetes&              // 全文搜索
    tags=k8s,devops&                // 按标签（metadata 中）
    limit=100&
    offset=0
```

### 分析 API

```go
// GET /api/v1/stats - 统计概览
// GET /api/v1/stats/trends?type=decision&period=month - 趋势分析
// GET /api/v1/stats/sources - 按来源统计
```

---

## 与 v2 的差异

| 维度 | v2 | v3 |
|------|----|----|
| **定位** | News Agent + Tech-Doc Agent | 个人数据中枢 |
| **类型** | 枚举 (brief/news/tracking/insight/log) | 开放（任意字符串） |
| **来源** | agent 字段 | source_system + source_agent |
| **租户** | 无 | 无（个人使用） |
| **元数据** | source_metadata | metadata |
| **API** | Agent 专用 | 统一 CRUD |
| **扩展性** | 受限于枚举 | 完全开放 |

---

## 迁移路径

### 从 v2 迁移到 v3

```sql
-- 1. 添加新字段
ALTER TABLE items ADD COLUMN source_system VARCHAR(64);
ALTER TABLE items ADD COLUMN source_agent VARCHAR(64);
ALTER TABLE items ADD COLUMN category VARCHAR(64);
ALTER TABLE items ADD COLUMN summary TEXT;

-- 2. 数据迁移
UPDATE items SET 
    source_system = agent,
    source_agent = agent,
    metadata = source_metadata;

-- 3. 重命名字段
ALTER TABLE items RENAME COLUMN source_metadata TO metadata;

-- 4. 删除约束
ALTER TABLE items DROP CONSTRAINT items_type_check;

-- 5. 添加新索引
-- (见上方索引设计)
```

---

## 未来扩展

### 向量搜索 (pgvector)

```sql
CREATE EXTENSION IF NOT EXISTS vector;

ALTER TABLE items ADD COLUMN embedding vector(1536);

CREATE INDEX idx_items_embedding ON items 
USING ivfflat (embedding vector_cosine_ops)
WITH (lists = 100);

-- 相似度查询
SELECT id, title, 
       1 - (embedding <=> query_vector) AS similarity
FROM items
ORDER BY embedding <=> query_vector
LIMIT 10;
```

### 知识图谱（预留）

```sql
CREATE TABLE relations (
    id          SERIAL PRIMARY KEY,
    from_item   VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    to_item     VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    relation    VARCHAR(64) NOT NULL,  -- relates_to, depends_on, causes, etc.
    metadata    JSONB DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(from_item, to_item, relation)
);

CREATE INDEX idx_relations_from ON relations(from_item);
CREATE INDEX idx_relations_to ON relations(to_item);
CREATE INDEX idx_relations_type ON relations(relation);
```

---

## 决策记录

### 为什么不用 namespace？

**原因：个人使用，无需多租户隔离。**

- 只有一个用户（kitsnail）
- 所有数据都是个人的
- 不需要用户级别的隔离
- 用 `source_system` 区分来源即可

### 为什么开放 type？

**原因：AI 工具多样，数据类型不可预知。**

- 不同工具产生的数据类型不同
- 强制枚举会限制扩展性
- 开放类型 + 约定优于配置

### 为什么用 source_system 而不是 agent？

**原因：更准确描述数据来源。**

- `source_system` = 哪个系统（openclaw, picoclaw）
- `source_agent` = 系统内的哪个 agent（可选）
- 更灵活，更清晰

---

## 下一步

1. **确认设计** - 是否有遗漏或需要调整
2. **实现迁移** - v2 → v3 迁移脚本
3. **更新 API** - 实现统一 CRUD
4. **集成测试** - 各 AI 工具接入验证

---

## 附录：完整 Schema

```sql
-- ============================================
-- Insight Hub Schema v3
-- ============================================

-- 扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
-- CREATE EXTENSION IF NOT EXISTS "vector";  -- 未来

-- 核心表
CREATE TABLE items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(64) NOT NULL,
    category        VARCHAR(64),
    source_system   VARCHAR(64) NOT NULL,
    source_agent    VARCHAR(64),
    source_url      TEXT,
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    summary         TEXT,
    metadata        JSONB DEFAULT '{}',
    status          VARCHAR(32) DEFAULT 'active',
    priority        INTEGER DEFAULT 0,
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    search_vector   TSVECTOR
);

-- 索引（见上方设计）

-- 触发器：自动更新 updated_at
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

-- 触发器：自动更新 search_vector
CREATE OR REPLACE FUNCTION update_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector = to_tsvector('english',
        coalesce(NEW.title, '') || ' ' ||
        coalesce(NEW.content, '') || ' ' ||
        coalesce(NEW.summary, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER items_search_vector
    BEFORE INSERT OR UPDATE ON items
    FOR EACH ROW
    EXECUTE FUNCTION update_search_vector();

-- 标签表（可选，如果需要独立标签管理）
CREATE TABLE tags (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(128) NOT NULL UNIQUE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE item_tags (
    item_id     VARCHAR(64) REFERENCES items(id) ON DELETE CASCADE,
    tag_id      INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (item_id, tag_id)
);

-- 来源系统注册表（可选）
CREATE TABLE source_systems (
    name        VARCHAR(64) PRIMARY KEY,
    description TEXT,
    config      JSONB DEFAULT '{}',
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 初始化来源系统
INSERT INTO source_systems (name, description) VALUES
    ('openclaw', 'openclaw AI system'),
    ('picoclaw', 'picoclaw AI system'),
    ('opencode', 'opencode development tool'),
    ('claude-code', 'Claude Code assistant'),
    ('codex', 'OpenAI Codex'),
    ('manual', 'Manual entry'),
    ('web', 'Web scraper')
ON CONFLICT (name) DO NOTHING;
```
