# Insight Hub 数据模型 v2 迁移指南

> **日期**: 2026-03-01
> **版本**: v2.0.0

---

## 概述

本次迁移将数据模型从 v1 升级到 v2，主要变更：

1. **type 枚举更新**: `bookmark/note/chat/code/task` → `brief/news/tracking/insight/log`
2. **新增 agent 字段**: 支持多 Agent 场景
3. **source_metadata 结构化**: 定义了每种类型的元数据规范
4. **追踪历史限制**: history 数组最多 30 条

---

## 迁移步骤

### 1. 备份数据

```bash
# 执行迁移脚本会自动备份
./scripts/migrate-v2.sh

# 或手动备份
docker exec insight-hub-postgres pg_dump -U insight insight_hub > backup.sql
```

### 2. 执行迁移

```bash
# 方式 1: 使用迁移脚本（推荐）
chmod +x scripts/migrate-v2.sh
./scripts/migrate-v2.sh

# 方式 2: 使用 Go 迁移（自动执行）
# 启动服务时会自动执行迁移
go run cmd/server/main.go

# 方式 3: 手动执行 SQL
cat internal/repository/migrations/v2_data_model.sql | docker exec -i insight-hub-postgres psql -U insight -d insight_hub
```

### 3. 验证迁移

```bash
# 运行验证脚本
chmod +x scripts/verify-db-v2.sh
./scripts/verify-db-v2.sh
```

### 4. 运行测试

```bash
# 单元测试
go test ./internal/repository -v -short

# 集成测试（需要数据库连接）
go test ./internal/repository -v
```

---

## 数据映射

### type 映射

| v1 | v2 | 说明 |
|----|----|----|
| bookmark | news | 书签 → 新闻 |
| note | insight | 笔记 → 洞察 |
| chat | log | 聊天 → 日志 |
| code | log | 代码 → 日志 |
| task | log | 任务 → 日志 |

### agent 推断

根据 source 字段推断 agent：

| source 模式 | agent |
|------------|-------|
| 包含 news/rss/feed | news-agent |
| 包含 github/doc/tech | tech-doc-agent |
| 其他 | unknown |

---

## 回滚

如果迁移出现问题，可以回滚：

```bash
# 使用备份恢复
cat /tmp/insight_hub_backup_*.sql | docker exec -i insight-hub-postgres psql -U insight -d insight_hub
```

---

## 新增功能

### 1. QueryByFilter - 高级查询

```go
filter := repository.QueryFilter{
    Type:     "news",
    Agent:    "news-agent",
    DateFrom: &startTime,
    DateTo:   &endTime,
    Search:   "GPT",
    Limit:    50,
}
items, total, err := itemRepo.QueryByFilter(ctx, filter)
```

### 2. UpdateTrackingHistory - 更新追踪历史

```go
historyItem := model.TrackingHistoryItem{
    Date:        "2026-03-01",
    Summary:     "OpenAI 发布 GPT-5",
    Importance:  0.9,
    SourceCount: 5,
}
err := itemRepo.UpdateTrackingHistory(ctx, itemID, historyItem)
```

### 3. 视图

```sql
-- 活跃追踪对象
SELECT * FROM active_trackings;

-- 最近简报
SELECT * FROM recent_briefs;
```

---

## 验收标准

- [x] 数据库 Schema 更新完成
- [x] Go 结构体与 Schema 对齐
- [x] 单元测试覆盖所有类型
- [ ] JSONB 查询性能 < 100ms（10 万条数据）
- [ ] history 数组长度限制生效
- [ ] API 文档更新

---

## 注意事项

1. **数据迁移不可逆**: v1 的 type 值会被映射到 v2，无法完全恢复
2. **agent 默认值**: 现有数据的 agent 会根据 source 推断，可能不准确
3. **测试环境**: 本次迁移在测试环境执行，生产环境需要更严格的流程

---

## 相关文档

- [数据模型设计 v2](./data-model-v2.md)
- [完整方案](./plan.md)
- [设计复盘](./design-review.md)
