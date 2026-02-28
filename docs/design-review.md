# Insight Hub 设计复盘

> **目的**：从实际使用场景出发，重新审视系统设计
> **创建时间**：2026-03-01

---

## 1. 原始需求回顾

### 1.1 News Agent 的需求

**定时任务**：
- 每日简报 @ 8:00 Beijing → Discord #news

**数据存储需求**：
1. **已发送简报记录** — 用于去重、回溯
2. **追踪话题列表** — 6个活跃话题，每日检查进展
3. **去重记录** — 7天内已推送的新闻 URL
4. **信源评估** — 记录哪些信源更可靠
5. **高价值洞察** — 积累有深度的分析判断

**使用场景**：
```
每日简报流程：
1. 读取追踪话题列表 → 知道要关注什么
2. 搜索各话题的最新进展 → 获取新闻
3. 检查去重记录 → 过滤已推送的新闻
4. 生成简报 → 推送到 Discord
5. 记录已发送简报 → 写入 items (type=brief)
6. 记录新闻 URL → 写入 dedup
7. 更新追踪话题进展 → 更新 items (type=tracking)
```

### 1.2 Tech-Doc Agent 的需求

**定时任务**：
- 每日技术简报 @ 07:30

**数据存储需求**：
1. **问题库** — 长期追踪的行业问题（深度）
2. **趋势图谱** — 趋势演变、信号追踪
3. **洞察笔记** — 深度分析、模式发现
4. **追踪话题** — 轻量级追踪（只需知道有无进展）
5. **迭代日志** — 简报质量优化记录

**使用场景**：
```
每日简报流程：
1. 读取问题库 → 知道要探索什么问题
2. 读取趋势图谱 → 知道当前趋势状态
3. 搜索技术动态 → 获取信息
4. 生成简报（问题驱动）→ 推送
5. 记录简报 → 写入 items (type=brief)
6. 更新趋势图谱 → 更新 items (type=insight)
7. 发现新问题 → 写入问题库
```

---

## 2. 当前设计的问题

### 2.1 数据模型不够清晰

**问题**：`items.type` 的定义模糊

当前定义：`brief, news, tracking, insight, log`

**疑问**：
- `brief` 和 `news` 的区别是什么？
- 简报本身应该是什么类型？
- 简报中包含的多条新闻，应该逐条存储还是整体存储？

**实际需求**：
- News Agent：简报是"一次推送的多条新闻集合"
- Tech-Doc Agent：简报是"围绕问题的技术洞察"

### 2.2 追踪话题的存储方式不明确

**问题**：tracking 类型的 item 应该如何存储？

当前设计：
```sql
type = 'tracking'
source_metadata->>'track_id' = 'T001'
source_metadata->>'track_type' = '问题追踪'
source_metadata->>'status' = 'active'
```

**疑问**：
- 追踪话题的"最新进展"如何更新？是更新原记录，还是追加新记录？
- 追踪历史如何保留？
- 如何区分"追踪话题定义"和"追踪进展记录"？

### 2.3 多 Agent 场景考虑不足

**问题**：虽然设计了 `agent` 字段，但未明确多 Agent 的协作模式

**场景**：
- News Agent 和 Tech-Doc Agent 可能关注同一话题（如 AI 发展趋势）
- 如何避免重复存储？
- 如何共享洞察？

---

## 3. 重新设计：从使用场景出发

### 3.1 核心概念重新定义

#### 概念 1：简报（Brief）

**定义**：一次推送的内容集合

**特征**：
- 由某个 agent 生成
- 包含多条内容（新闻、洞察、追踪更新等）
- 有明确的发送时间和渠道
- 是"推送事件"，不是"内容本身"

**存储方式**：
```json
{
  "id": "brief-20260301-news",
  "type": "brief",
  "agent": "news",
  "title": "每日简报 2026-03-01",
  "content": "## 头条\n...\n## 追踪更新\n...",
  "source_metadata": {
    "sent_to": "discord#news",
    "sent_at": "2026-03-01T08:00:00+08:00",
    "item_count": 5,
    "topics": ["T001", "T002"]
  }
}
```

#### 概念 2：新闻条目（News Item）

**定义**：单条新闻/信息的原始记录

**特征**：
- 来自某个信源
- 有 URL（可去重）
- 可被多个简报引用
- 是"内容本身"

**存储方式**：
```json
{
  "id": "news-abc123",
  "type": "news",
  "agent": "news",
  "title": "Nvidia H200 获准向中国出货",
  "content": "据 CNBC 报道...",
  "source": "CNBC",
  "source_url": "https://cnbc.com/...",
  "source_metadata": {
    "published_at": "2026-02-26T10:00:00Z",
    "category": "semiconductor",
    "importance": "high"
  }
}
```

#### 概念 3：追踪话题（Tracking Topic）

**定义**：持续关注的话题/问题

**特征**：
- 有明确的追踪 ID
- 有状态（active/paused/archived）
- 有关键词列表
- 有历史进展记录

**存储方式（方案 A：单记录更新）**：
```json
{
  "id": "T001",
  "type": "tracking",
  "agent": "news",
  "title": "半导体出口管制",
  "source_metadata": {
    "keywords": ["Nvidia H200", "AMD MI325", "BIS"],
    "status": "active",
    "started_at": "2026-02-28",
    "last_check": "2026-03-01",
    "latest_progress": "2月26日CNBC报道...",
    "history": [
      {
        "date": "2026-02-28",
        "progress": "开始追踪"
      },
      {
        "date": "2026-03-01",
        "progress": "2月26日CNBC报道..."
      }
    ]
  }
}
```

**存储方式（方案 B：主记录 + 进展记录）**：
```json
// 主记录
{
  "id": "T001",
  "type": "tracking",
  "agent": "news",
  "title": "半导体出口管制",
  "source_metadata": {
    "keywords": ["Nvidia H200", "AMD MI325", "BIS"],
    "status": "active",
    "started_at": "2026-02-28"
  }
}

// 进展记录（追加）
{
  "id": "T001-progress-20260301",
  "type": "tracking_progress",
  "agent": "news",
  "title": "T001 进展 2026-03-01",
  "source_metadata": {
    "track_id": "T001",
    "progress": "2月26日CNBC报道..."
  },
  "occurred_at": "2026-03-01"
}
```

#### 概念 4：洞察（Insight）

**定义**：深度分析、模式发现、价值判断

**特征**：
- 非时效性
- 可跨话题关联
- 可被多个简报引用

**存储方式**：
```json
{
  "id": "insight-xyz",
  "type": "insight",
  "agent": "tech-doc",
  "title": "AI 推理系统演进趋势",
  "content": "从 2024-2026 的技术演进来看...",
  "source_metadata": {
    "related_questions": ["Q1"],
    "related_trends": ["T001"],
    "confidence": "high"
  }
}
```

### 3.2 数据模型重新设计

#### 方案 A：保持单表，明确类型语义

```sql
-- items 表保持不变，但重新定义 type 的语义
CREATE TABLE items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(32) NOT NULL,  -- brief, news, tracking, tracking_progress, insight, question, trend
    agent           VARCHAR(32) NOT NULL,
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    source          VARCHAR(64),
    source_url      TEXT,
    source_metadata JSONB DEFAULT '{}',
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT items_type_check CHECK (
        type IN ('brief', 'news', 'tracking', 'tracking_progress', 'insight', 'question', 'trend', 'log')
    )
);
```

**优点**：
- 保持简单
- JSONB 灵活存储各种元数据

**缺点**：
- 类型语义仍然模糊
- 查询时需要依赖 source_metadata

#### 方案 B：多表设计，明确业务语义

```sql
-- 简报表
CREATE TABLE briefs (
    id              VARCHAR(64) PRIMARY KEY,
    agent           VARCHAR(32) NOT NULL,
    title           VARCHAR(512) NOT NULL,
    content         TEXT NOT NULL,
    sent_to         VARCHAR(128),
    sent_at         TIMESTAMP WITH TIME ZONE,
    item_count      INTEGER DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 新闻表
CREATE TABLE news (
    id              VARCHAR(64) PRIMARY KEY,
    agent           VARCHAR(32) NOT NULL,
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    source          VARCHAR(64),
    source_url      TEXT UNIQUE,
    published_at    TIMESTAMP WITH TIME ZONE,
    category        VARCHAR(64),
    importance      VARCHAR(16),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 追踪话题表
CREATE TABLE trackings (
    id              VARCHAR(32) PRIMARY KEY,  -- T001, T002
    agent           VARCHAR(32) NOT NULL,
    title           VARCHAR(512) NOT NULL,
    keywords        TEXT[] NOT NULL,
    status          VARCHAR(16) DEFAULT 'active',
    started_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    last_check      TIMESTAMP WITH TIME ZONE,
    latest_progress TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 追踪进展表
CREATE TABLE tracking_progress (
    id              SERIAL PRIMARY KEY,
    tracking_id     VARCHAR(32) REFERENCES trackings(id) ON DELETE CASCADE,
    progress        TEXT NOT NULL,
    occurred_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 洞察表
CREATE TABLE insights (
    id              VARCHAR(64) PRIMARY KEY,
    agent           VARCHAR(32) NOT NULL,
    title           VARCHAR(512) NOT NULL,
    content         TEXT NOT NULL,
    confidence      VARCHAR(16),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 问题表（Tech-Doc 专用）
CREATE TABLE questions (
    id              VARCHAR(32) PRIMARY KEY,  -- Q1, Q2
    agent           VARCHAR(32) NOT NULL,
    question        TEXT NOT NULL,
    status          VARCHAR(16) DEFAULT 'active',
    context         TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 去重表（保持不变）
CREATE TABLE dedup (
    id              SERIAL PRIMARY KEY,
    url_hash        VARCHAR(128) NOT NULL UNIQUE,
    url             TEXT NOT NULL,
    news_id         VARCHAR(64) REFERENCES news(id) ON DELETE SET NULL,
    agent           VARCHAR(32) NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at      TIMESTAMP WITH TIME ZONE NOT NULL
);
```

**优点**：
- 业务语义清晰
- 查询简单直接
- 类型安全

**缺点**：
- 表数量增加
- 需要更多代码

---

## 4. 决策点

### 4.1 数据模型选择

**方案 A（单表 + JSONB）** vs **方案 B（多表）**

| 维度 | 方案 A | 方案 B |
|------|--------|--------|
| 灵活性 | ✅ 高 | ⚠️ 需要改表 |
| 查询复杂度 | ⚠️ 需要 JSONB 查询 | ✅ 简单直接 |
| 类型安全 | ⚠️ 运行时检查 | ✅ 编译时检查 |
| 代码量 | ✅ 少 | ⚠️ 多 |
| 适用场景 | 快速迭代、不确定需求 | 需求稳定、业务复杂 |

**建议**：考虑到 Insight Hub 是新系统，需求还在探索，**先采用方案 A（单表 + JSONB）**，但：
1. 明确 type 的语义定义
2. 为每种 type 定义 source_metadata 的 schema
3. 未来如果某种 type 的查询成为瓶颈，再拆分

### 4.2 追踪话题的存储方式

**方案 A（单记录更新）** vs **方案 B（主记录 + 进展记录）**

| 维度 | 方案 A | 方案 B |
|------|--------|--------|
| 简单性 | ✅ 简单 | ⚠️ 需要关联 |
| 历史追溯 | ⚠️ JSONB 数组 | ✅ 独立记录 |
| 查询性能 | ✅ 单表 | ⚠️ 需要关联 |
| 适用场景 | 进展频率低 | 进展频率高、需要历史 |

**建议**：考虑到追踪话题的进展频率（每日一次），**先采用方案 A（单记录更新）**，但：
1. 在 source_metadata 中维护 history 数组
2. 限制 history 数组长度（如最近 30 条）
3. 超过限制时归档到单独的文件或表

### 4.3 简报与新闻的关系

**问题**：简报包含多条新闻，如何关联？

**方案 A**：简报的 content 字段直接存储完整内容（当前设计）

**方案 B**：简报与新闻通过关联表关联

```sql
CREATE TABLE brief_news (
    brief_id    VARCHAR(64) REFERENCES briefs(id) ON DELETE CASCADE,
    news_id     VARCHAR(64) REFERENCES news(id) ON DELETE CASCADE,
    position    INTEGER,
    PRIMARY KEY (brief_id, news_id)
);
```

**建议**：**先采用方案 A**，因为：
1. 简报是"推送事件"，不需要复杂的关联查询
2. 简报生成时已经确定了内容，后续不需要修改
3. 如果需要回溯某条新闻，可以通过去重表反查

---

## 5. 修订后的数据模型

### 5.1 Items 表（修订版）

```sql
CREATE TABLE items (
    id              VARCHAR(64) PRIMARY KEY,
    type            VARCHAR(32) NOT NULL,
    agent           VARCHAR(32) NOT NULL,
    title           VARCHAR(512) NOT NULL,
    content         TEXT,
    source          VARCHAR(64),
    source_url      TEXT,
    source_metadata JSONB DEFAULT '{}',
    occurred_at     TIMESTAMP WITH TIME ZONE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT items_type_check CHECK (
        type IN ('brief', 'news', 'tracking', 'insight', 'question', 'trend', 'log')
    )
);

-- 索引保持不变
CREATE INDEX idx_items_type ON items(type);
CREATE INDEX idx_items_agent ON items(agent);
CREATE INDEX idx_items_occurred_at ON items(occurred_at DESC);
CREATE INDEX idx_items_created_at ON items(created_at DESC);
CREATE INDEX idx_items_source ON items(source);
CREATE INDEX idx_items_metadata ON items USING GIN (source_metadata);
```

### 5.2 Type 语义定义

| Type | 含义 | 典型字段 | 示例 |
|------|------|---------|------|
| `brief` | 简报（推送事件） | sent_to, sent_at, item_count | News Agent 每日简报 |
| `news` | 新闻条目 | source, source_url, published_at, category | 单条新闻记录 |
| `tracking` | 追踪话题 | keywords, status, history | 半导体出口管制追踪 |
| `insight` | 洞察笔记 | confidence, related_questions | AI 推理系统演进趋势 |
| `question` | 问题（深度追踪） | context, status | 如何设计高可用 K8s 集群？ |
| `trend` | 趋势 | signals, evolution | AI 从大模型到智能体 |
| `log` | 日志/事件 | event_type, details | 系统操作日志 |

### 5.3 Source Metadata Schema（每种 Type）

#### brief

```json
{
  "sent_to": "discord#news",
  "sent_at": "2026-03-01T08:00:00+08:00",
  "item_count": 5,
  "topics": ["T001", "T002"],
  "channels": ["discord", "telegram"]
}
```

#### news

```json
{
  "published_at": "2026-02-26T10:00:00Z",
  "category": "semiconductor",
  "importance": "high",
  "related_tracking": ["T001"]
}
```

#### tracking

```json
{
  "keywords": ["Nvidia H200", "AMD MI325", "BIS"],
  "status": "active",
  "started_at": "2026-02-28",
  "last_check": "2026-03-01",
  "latest_progress": "2月26日CNBC报道...",
  "history": [
    {
      "date": "2026-02-28",
      "progress": "开始追踪"
    },
    {
      "date": "2026-03-01",
      "progress": "2月26日CNBC报道..."
    }
  ]
}
```

#### insight

```json
{
  "confidence": "high",
  "related_questions": ["Q1"],
  "related_trackings": ["T001"],
  "related_trends": ["trend-001"]
}
```

#### question

```json
{
  "context": "在 K8s 集群中，节点就绪状态检测...",
  "status": "active",
  "created_from": "daily-brief",
  "related_trends": ["trend-001"]
}
```

#### trend

```json
{
  "signals": [
    {
      "date": "2026-02-01",
      "signal": "OpenAI 发布 GPT-5"
    }
  ],
  "evolution": "从大模型竞赛到智能体应用",
  "status": "active"
}
```

---

## 6. API 设计（修订版）

### 6.1 核心 API

```
# Items CRUD
POST   /api/items              # 创建 item
GET    /api/items/:id          # 获取单个 item
PUT    /api/items/:id          # 更新 item
DELETE /api/items/:id          # 删除 item
GET    /api/items              # 查询 items（支持筛选）

# 查询参数
?type=brief&agent=news&date_from=2026-03-01&date_to=2026-03-01
?type=tracking&status=active
?type=news&category=semiconductor

# 去重
POST   /api/dedup/check        # 检查是否重复
POST   /api/dedup              # 记录去重
DELETE /api/dedup/expired      # 清理过期记录

# 视图
GET    /api/trackings/active   # 活跃追踪话题
GET    /api/briefs/recent      # 最近简报
GET    /api/questions/active   # 活跃问题
GET    /api/trends/active      # 活跃趋势
```

### 6.2 Agent 专用 API

```
# News Agent
POST   /api/news/collect       # 批量收集新闻（自动去重）
POST   /api/briefs/generate    # 生成简报（News Agent 专用）

# Tech-Doc Agent
POST   /api/questions          # 创建问题
POST   /api/trends             # 更新趋势
GET    /api/questions/:id/insights  # 获取问题相关的洞察
```

---

## 7. 下一步

### 7.1 需要确认的问题

1. **数据模型**：方案 A（单表）还是方案 B（多表）？
2. **追踪话题存储**：方案 A（单记录更新）还是方案 B（主记录 + 进展记录）？
3. **简报与新闻关系**：是否需要关联表？
4. **多 Agent 协作**：是否需要共享数据？如何避免重复？

### 7.2 等待用户确认后

1. 更新 `docs/plan.md`
2. 编写数据库迁移脚本
3. 重构 Phase 2 实现
4. 继续 Phase 3 API 实现

---

_最后更新：2026-03-01_
