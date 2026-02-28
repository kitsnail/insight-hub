# Insight Hub 数据模型设计 v2

> **状态**: 最终方案
> **日期**: 2026-03-01
> **决策**: 单表 + JSONB，预留扩展空间

---

## 核心决策

| 决策点 | 方案 | 理由 |
|--------|------|------|
| 数据模型 | 单表 + JSONB | 需求还在探索，快速迭代优先 |
| 追踪话题 | 单记录 + history 数组 | 简单查询，限制 30 条 |
| 简报与新闻 | 独立，无关联 | 两个独立任务 |
| 多 Agent | 各自独立 | 通过 `agent` 字段区分 |

---

## 数据模型

### items 表

```sql
CREATE TABLE items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(32) NOT NULL,           -- brief, news, tracking, insight, log
    agent           VARCHAR(32) NOT NULL,           -- news-agent, tech-doc-agent, ...
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    source          VARCHAR(64),                    -- 来源标识
    source_url      TEXT,                           -- 原始 URL
    source_metadata JSONB DEFAULT '{}',             -- 灵活元数据
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT items_type_check CHECK (type IN ('brief', 'news', 'tracking', 'insight', 'log'))
);
```

### 索引策略

```sql
-- 高频查询字段（原生列）
CREATE INDEX idx_items_type ON items(type);
CREATE INDEX idx_items_agent ON items(agent);
CREATE INDEX idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX idx_items_created_at ON items(created_at DESC);
CREATE INDEX idx_items_source ON items(source);

-- JSONB 字段（GIN 索引）
CREATE INDEX idx_items_metadata ON items USING GIN (source_metadata);

-- 全文搜索
CREATE INDEX idx_items_fulltext ON items 
USING GIN (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(content, '')));

-- 模糊搜索
CREATE INDEX idx_items_title_trgm ON items USING GIN (title gin_trgm_ops);
```

---

## source_metadata 字段规范

### 通用字段

所有类型都可以包含：

```json
{
  "raw_data": {},           // 原始数据（可选）
  "extra": {}               // 其他扩展字段
}
```

### 按类型的特定字段

#### type = 'news'

```json
{
  "author": "张三",
  "published_at": "2026-03-01T10:00:00Z",
  "category": "AI",
  "keywords": ["GPT", "LLM"],
  "summary": "文章摘要...",
  "sentiment": "positive",
  "importance": 0.8
}
```

#### type = 'brief'

```json
{
  "brief_type": "daily",        // daily, weekly, breaking
  "period_start": "2026-03-01",
  "period_end": "2026-03-01",
  "sent_to": ["telegram", "email"],
  "sent_at": "2026-03-01T08:00:00Z",
  "news_count": 15,
  "topics": ["AI", "Cloud Native"]
}
```

#### type = 'tracking'

```json
{
  "track_id": "track-001",
  "track_type": "trend",        // trend, issue, topic
  "topic": "AI 进展",
  "status": "active",           // active, paused, completed
  "last_check": "2026-03-01T00:00:00Z",
  "check_interval": "24h",
  "history": [
    {
      "date": "2026-03-01",
      "summary": "OpenAI 发布 GPT-5...",
      "importance": 0.9,
      "source_count": 5
    },
    {
      "date": "2026-02-28",
      "summary": "Claude 3.5 发布...",
      "importance": 0.8,
      "source_count": 3
    }
  ]
}
```

**限制**: `history` 数组最多保留 **30 条**，超出时：
- 删除最旧的记录
- 或归档到对象存储（未来功能）

#### type = 'insight'

```json
{
  "insight_type": "observation",  // observation, prediction, question
  "confidence": 0.8,
  "related_items": ["item-id-1", "item-id-2"],
  "actionable": true,
  "action_taken": "已通知团队"
}
```

#### type = 'log'

```json
{
  "log_type": "agent_action",     // agent_action, system, error
  "action": "fetch_news",
  "result": "success",
  "duration_ms": 1234,
  "error": null
}
```

---

## 风险与缓解

### 风险 1: 查询性能

**问题**: JSONB 字段查询比原生列慢

**缓解**:
- 高频查询字段用原生列（`type`, `agent`, `occurred_at`）
- 对 JSONB 字段建 GIN 索引
- 监控慢查询，必要时提取为原生列

### 风险 2: 类型安全

**问题**: 应用层需要手动校验 JSON 结构

**缓解**:
- 在 Repository 层做类型转换
- 使用 Go 结构体映射
- 编写单元测试覆盖

### 风险 3: 迁移成本

**问题**: 未来拆分多表需要数据迁移

**缓解**:
- 预留 `item_type` 字段（已存在 `type` 字段）
- 设计迁移脚本模板
- 保持 JSONB 字段结构清晰

### 风险 4: 索引膨胀

**问题**: JSONB 索引占用空间较大

**缓解**:
- 只对必要字段建索引
- 定期监控索引大小
- 使用部分索引（Partial Index）减少大小

---

## 扩展路径

### 阶段 1: MVP（当前）

- 单表 + JSONB
- history 数组限制 30 条
- 各 Agent 独立存储

### 阶段 2: 性能优化（未来）

- 提取高频 JSONB 字段为原生列
- 添加缓存层
- 优化查询语句

### 阶段 3: 数据拆分（未来）

当单表数据量超过 100 万或查询性能下降时：

```
items (单表)
    ↓ 拆分
├── news (新闻表)
├── briefs (简报表)
├── trackings (追踪表)
├── insights (洞察表)
└── logs (日志表)
```

**迁移策略**:
1. 创建新表
2. 双写（新旧表同时写入）
3. 历史数据迁移
4. 切换读取
5. 删除旧表

### 阶段 4: 多 Agent 共享（未来）

当需要跨 Agent 查询时：

```sql
-- 添加共享字段
ALTER TABLE items ADD COLUMN is_shared BOOLEAN DEFAULT FALSE;
ALTER TABLE items ADD COLUMN shared_with TEXT[];

-- 创建共享视图
CREATE VIEW shared_items AS
SELECT * FROM items WHERE is_shared = TRUE;
```

---

## 与 v1 的差异

| 项目 | v1 (当前实现) | v2 (本方案) |
|------|-------------|------------|
| type 枚举 | bookmark, note, chat, code, task | **brief, news, tracking, insight, log** |
| agent 字段 | ❌ 缺失 | ✅ 必填 |
| source_metadata | 简单 JSON | **结构化规范** |
| history 数组 | 无限制 | **限制 30 条** |
| 扩展路径 | 无 | **明确规划** |

---

## 实施步骤

### Step 1: 更新数据库 Schema

```sql
-- 1. 修改 type 约束
ALTER TABLE items DROP CONSTRAINT items_type_check;
ALTER TABLE items ADD CONSTRAINT items_type_check 
  CHECK (type IN ('brief', 'news', 'tracking', 'insight', 'log'));

-- 2. 添加 agent 字段（如果不存在）
ALTER TABLE items ADD COLUMN IF NOT EXISTS agent VARCHAR(32) NOT NULL DEFAULT 'unknown';

-- 3. 更新索引
CREATE INDEX IF NOT EXISTS idx_items_agent ON items(agent);
```

### Step 2: 更新 Go 结构体

```go
type Item struct {
    ID             string                 `json:"id"`
    Type           string                 `json:"type"`      // brief, news, tracking, insight, log
    Agent          string                 `json:"agent"`     // news-agent, tech-doc-agent
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

// 类型常量
const (
    ItemTypeBrief    = "brief"
    ItemTypeNews     = "news"
    ItemTypeTracking = "tracking"
    ItemTypeInsight  = "insight"
    ItemTypeLog      = "log"
)

// Agent 常量
const (
    AgentNews    = "news-agent"
    AgentTechDoc = "tech-doc-agent"
)
```

### Step 3: 数据迁移

如果已有数据，需要迁移：

```sql
-- 示例：将旧 type 映射到新 type
UPDATE items SET type = 'news' WHERE type = 'bookmark';
UPDATE items SET type = 'insight' WHERE type = 'note';
-- ... 其他映射
```

---

## 验收标准

- [ ] 数据库 Schema 更新完成
- [ ] Go 结构体与 Schema 对齐
- [ ] 单元测试覆盖所有类型
- [ ] JSONB 查询性能 < 100ms（10 万条数据）
- [ ] history 数组长度限制生效
- [ ] API 文档更新

---

## 参考资料

- [PostgreSQL JSONB 最佳实践](https://www.postgresql.org/docs/current/datatype-json.html)
- [pg_trgm 扩展文档](https://www.postgresql.org/docs/current/pgtrgm.html)
- [Insight Hub 原始需求](./design-review.md)
