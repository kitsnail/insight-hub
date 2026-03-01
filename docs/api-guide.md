# Insight Hub API 使用指南

> **版本:** v3  
> **基础 URL:** `http://localhost:8090/api/v1`  
> **认证:** 内网无认证

---

## 快速开始

### 健康检查

```bash
curl http://localhost:8090/api/v1/health
```

响应：
```json
{
  "status": "ok",
  "version": "v3"
}
```

---

## 核心 API

### 1. 创建 Item

**POST** `/api/v1/items`

```bash
curl -X POST http://localhost:8090/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "type": "news",
    "source_system": "picoclaw",
    "source_agent": "news-agent",
    "title": "Kubernetes 1.30 发布",
    "content": "Kubernetes 1.30 引入了多项新特性...",
    "metadata": {
      "author": "CNCF",
      "importance": 0.8
    },
    "tags": ["k8s", "release"]
  }'
```

**请求体字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `id` | string | 否 | 主键，不提供则自动生成 `{source_system}:{uuid}` |
| `type` | string | ✅ | 数据类型（开放，如 `news`, `memory`, `decision`） |
| `category` | string | 否 | 二级分类 |
| `source_system` | string | ✅ | 来源系统（如 `picoclaw`, `opencode`） |
| `source_agent` | string | 否 | 来源 Agent（如 `news-agent`） |
| `source_url` | string | 否 | 原始 URL |
| `title` | string | ✅ | 标题 |
| `content` | string | 否 | 内容主体 |
| `summary` | string | 否 | 摘要 |
| `metadata` | object | 否 | 灵活元数据 |
| `status` | string | 否 | 状态，默认 `active` |
| `priority` | int | 否 | 优先级，默认 `0` |
| `occurred_at` | string | 否 | 事件发生时间（ISO8601） |
| `tags` | array | 否 | 标签名称列表 |

**响应：**
```json
{
  "id": "picoclaw:550e8400-e29b-41d4-a716-446655440000",
  "type": "news",
  "source_system": "picoclaw",
  "source_agent": "news-agent",
  "title": "Kubernetes 1.30 发布",
  "content": "Kubernetes 1.30 引入了多项新特性...",
  "metadata": {
    "author": "CNCF",
    "importance": 0.8
  },
  "status": "active",
  "priority": 0,
  "created_at": "2026-03-01T10:00:00Z",
  "updated_at": "2026-03-01T10:00:00Z",
  "tags": [
    {"id": 1, "name": "k8s"},
    {"id": 2, "name": "release"}
  ]
}
```

---

### 2. 获取单个 Item

**GET** `/api/v1/items/{id}`

```bash
curl http://localhost:8090/api/v1/items/picoclaw:550e8400-e29b-41d4-a716-446655440000
```

**响应：** 同创建响应

---

### 3. 列出 Items

**GET** `/api/v1/items`

```bash
# 基本查询
curl "http://localhost:8090/api/v1/items?limit=10"

# 按类型过滤
curl "http://localhost:8090/api/v1/items?type=news&limit=20"

# 按来源过滤
curl "http://localhost:8090/api/v1/items?source_system=picoclaw&type=memory"

# 全文搜索
curl "http://localhost:8090/api/v1/items?search=kubernetes&limit=10"

# 时间范围
curl "http://localhost:8090/api/v1/items?date_from=2026-03-01&date_to=2026-03-02"

# 标签过滤
curl "http://localhost:8090/api/v1/items?tags=k8s,devops"

# 排序
curl "http://localhost:8090/api/v1/items?sort_by=priority&sort_desc=true"
```

**查询参数：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `type` | string | - | 按类型过滤 |
| `category` | string | - | 按分类过滤 |
| `source_system` | string | - | 按来源系统过滤 |
| `source_agent` | string | - | 按来源 Agent 过滤 |
| `status` | string | `active` | 按状态过滤 |
| `date_from` | string | - | 起始日期（YYYY-MM-DD 或 ISO8601） |
| `date_to` | string | - | 结束日期 |
| `search` | string | - | 全文搜索（标题+内容+摘要） |
| `tags` | string | - | 标签过滤（逗号分隔） |
| `limit` | int | 100 | 每页数量（最大 1000） |
| `offset` | int | 0 | 偏移量 |
| `sort_by` | string | `created_at` | 排序字段 |
| `sort_desc` | bool | `true` | 是否降序 |

**响应：**
```json
{
  "items": [
    { /* Item 对象 */ },
    { /* Item 对象 */ }
  ],
  "total": 42,
  "limit": 10,
  "offset": 0,
  "has_more": true
}
```

---

### 4. 更新 Item

**PUT** `/api/v1/items/{id}`

```bash
curl -X PUT http://localhost:8090/api/v1/items/picoclaw:xxx \
  -H "Content-Type: application/json" \
  -d '{
    "title": "更新后的标题",
    "priority": 1,
    "status": "archived"
  }'
```

**请求体字段：** 所有字段可选，只更新提供的字段

---

### 5. 删除 Item

**DELETE** `/api/v1/items/{id}`

```bash
curl -X DELETE http://localhost:8090/api/v1/items/picoclaw:xxx
```

**注意：** 这是软删除，数据仍保留在数据库中，状态变为 `deleted`

---

## 辅助 API

### 去重检查

**POST** `/api/v1/items/check-duplicate`

```bash
curl -X POST http://localhost:8090/api/v1/items/check-duplicate \
  -H "Content-Type: application/json" \
  -d '{
    "source_system": "picoclaw",
    "source_agent": "news-agent",
    "external_id": "news-20260301-001",
    "title": "Kubernetes 1.30 发布"
  }'
```

**请求体字段：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `source_system` | string | ✅ | 来源系统 |
| `source_agent` | string | 否 | 来源 Agent |
| `external_id` | string | 否 | 外部系统唯一标识 |
| `title` | string | 否 | 标题匹配 |

**响应：**
```json
{
  "exists": true,
  "item_id": "picoclaw:existing-id",
  "title": "Kubernetes 1.30 发布",
  "occurred_at": "2026-03-01T08:00:00Z",
  "dedup_reason": "title"
}
```

---

### 批量创建

**POST** `/api/v1/items/batch`

```bash
curl -X POST http://localhost:8090/api/v1/items/batch \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {
        "type": "news",
        "source_system": "picoclaw",
        "title": "新闻 1"
      },
      {
        "type": "news",
        "source_system": "picoclaw",
        "title": "新闻 2"
      }
    ]
  }'
```

**限制：** 每批最多 1000 条

**响应：**
```json
{
  "total": 2,
  "succeeded": 2,
  "failed_count": 0,
  "created": [
    { /* Item 对象 */ },
    { /* Item 对象 */ }
  ],
  "failed": []
}
```

---

## 统计 API

### 统计概览

**GET** `/api/v1/stats`

```bash
curl http://localhost:8090/api/v1/stats
```

**响应：**
```json
{
  "total_items": 16,
  "by_type": {
    "article": 1,
    "memory": 1,
    "news": 7,
    "tracking": 6,
    "test": 1
  },
  "by_source": {
    "opencode": 1,
    "picoclaw": 14,
    "test": 1
  },
  "by_status": {
    "active": 16
  }
}
```

---

### 按类型统计

**GET** `/api/v1/stats/by-type`

```bash
curl http://localhost:8090/api/v1/stats/by-type
```

---

### 按来源统计

**GET** `/api/v1/stats/by-source`

```bash
curl http://localhost:8090/api/v1/stats/by-source
```

---

## 常见使用场景

### 场景 1：存储对话记忆

```bash
curl -X POST http://localhost:8090/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "type": "memory",
    "source_system": "picoclaw",
    "source_agent": "conversation",
    "title": "关于 Kubernetes 架构的讨论",
    "content": "用户询问了 K8s 的控制平面组件...",
    "metadata": {
      "session_id": "sess-20260301-001",
      "turn": 5,
      "tokens": 512
    },
    "tags": ["k8s", "architecture"]
  }'
```

---

### 场景 2：存储决策记录

```bash
curl -X POST http://localhost:8090/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "type": "decision",
    "source_system": "picoclaw",
    "title": "选择 PostgreSQL 作为主数据库",
    "content": "经过评估，决定使用 PostgreSQL...",
    "metadata": {
      "context": "Insight Hub 项目数据存储选型",
      "alternatives": [
        {"option": "MySQL", "reason": "JSONB 支持较弱"},
        {"option": "MongoDB", "reason": "不适合关系型数据"}
      ],
      "outcome": "PostgreSQL",
      "rationale": "支持 JSONB、全文搜索、扩展性好"
    },
    "priority": 1,
    "tags": ["architecture", "database"]
  }'
```

---

### 场景 3：存储追踪话题

```bash
curl -X POST http://localhost:8090/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "id": "picoclaw:tracking:k8s-security",
    "type": "tracking",
    "source_system": "picoclaw",
    "source_agent": "news-agent",
    "title": "T001: Kubernetes 安全动态",
    "content": "追踪 Kubernetes 安全相关的新闻和更新",
    "metadata": {
      "topic": "Kubernetes Security",
      "keywords": ["CVE", "security", "RBAC", "pod-security"],
      "history": []
    },
    "tags": ["tracking", "security", "k8s"]
  }'
```

---

### 场景 4：查询最近决策

```bash
curl "http://localhost:8090/api/v1/items?type=decision&limit=10&sort_by=occurred_at&sort_desc=true"
```

---

### 场景 5：全文搜索

```bash
curl "http://localhost:8090/api/v1/items?search=postgresql+jsonb&limit=20"
```

---

## 数据类型约定

### type 常见值

| type | 说明 | 典型 metadata |
|------|------|--------------|
| `memory` | 对话记忆 | `session_id`, `turn`, `tokens` |
| `note` | 笔记 | `tags`, `importance` |
| `news` | 新闻 | `author`, `published_at`, `sentiment` |
| `decision` | 决策记录 | `context`, `alternatives`, `outcome` |
| `task` | 任务 | `due`, `status`, `assignee` |
| `insight` | 洞察 | `confidence`, `related_items` |
| `tracking` | 追踪话题 | `topic`, `keywords`, `history` |
| `brief` | 简报 | `period`, `news_count` |
| `log` | 日志 | `action`, `result` |
| `code` | 代码片段 | `language`, `file` |
| `bookmark` | 书签 | `url`, `tags` |

### source_system 常见值

| source_system | 说明 |
|---------------|------|
| `openclaw` | openclaw AI 系统 |
| `picoclaw` | picoclaw AI 系统 |
| `opencode` | opencode 开发工具 |
| `claude-code` | Claude Code |
| `codex` | OpenAI Codex |
| `manual` | 手动录入 |
| `web` | Web 抓取 |

### status 值

| status | 说明 |
|--------|------|
| `active` | 活跃（默认） |
| `archived` | 已归档 |
| `deleted` | 已删除 |

---

## 错误响应

所有错误响应格式：

```json
{
  "error": "错误描述信息"
}
```

**常见 HTTP 状态码：**

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 最佳实践

### 1. ID 设计

推荐格式：`{source_system}:{type}:{uuid}` 或 `{source_system}:{uuid}`

```json
{
  "id": "picoclaw:news:abc123",
  "id": "opencode:decision:xyz789"
}
```

### 2. 时间字段

- `occurred_at`: 事件实际发生时间（如新闻发布时间）
- `created_at`: 数据入库时间（自动）
- `updated_at`: 最后更新时间（自动）

### 3. 元数据设计

保持结构一致，便于查询：

```json
{
  "metadata": {
    "importance": 0.8,           // 0-1 数值
    "confidence": 0.9,           // 0-1 数值
    "tags": ["k8s", "devops"],   // 标签数组
    "source": {                  // 来源详情
      "author": "CNCF",
      "url": "https://..."
    }
  }
}
```

### 4. 批量操作

大量数据写入使用批量 API：

```bash
# 推荐：批量创建
curl -X POST http://localhost:8090/api/v1/items/batch \
  -H "Content-Type: application/json" \
  -d '{"items": [...]}'

# 不推荐：循环单个创建
for item in items; do
  curl -X POST http://localhost:8090/api/v1/items -d "$item"
done
```

### 5. 去重策略

写入前先检查：

```bash
# 1. 检查是否已存在
curl -X POST http://localhost:8090/api/v1/items/check-duplicate \
  -d '{"source_system": "picoclaw", "external_id": "news-001"}'

# 2. 根据结果决定是否创建
```

---

## 附录：完整示例

### Python 客户端示例

```python
import requests

BASE_URL = "http://localhost:8090/api/v1"

def create_item(item_data):
    """创建 Item"""
    response = requests.post(f"{BASE_URL}/items", json=item_data)
    response.raise_for_status()
    return response.json()

def get_item(item_id):
    """获取 Item"""
    response = requests.get(f"{BASE_URL}/items/{item_id}")
    response.raise_for_status()
    return response.json()

def list_items(type=None, source_system=None, limit=100, offset=0):
    """列出 Items"""
    params = {"limit": limit, "offset": offset}
    if type:
        params["type"] = type
    if source_system:
        params["source_system"] = source_system
    
    response = requests.get(f"{BASE_URL}/items", params=params)
    response.raise_for_status()
    return response.json()

def search_items(query, limit=20):
    """全文搜索"""
    response = requests.get(f"{BASE_URL}/items", params={
        "search": query,
        "limit": limit
    })
    response.raise_for_status()
    return response.json()

# 使用示例
if __name__ == "__main__":
    # 创建
    item = create_item({
        "type": "news",
        "source_system": "picoclaw",
        "title": "Kubernetes 1.30 发布",
        "content": "新版本带来了多项改进...",
        "tags": ["k8s", "release"]
    })
    print(f"Created: {item['id']}")
    
    # 查询
    items = list_items(type="news", limit=10)
    print(f"Found {items['total']} news items")
    
    # 搜索
    results = search_items("kubernetes")
    print(f"Search results: {len(results['items'])} items")
```

---

## 更新记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-03-01 | v3.0.0 | 初始版本，v3 架构 |
