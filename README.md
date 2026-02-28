# Insight Hub

个人知识管理中枢 - 收集、整理、洞察。

## 功能特性

- 📥 **多源数据收集** - 浏览器书签、聊天记录、代码片段、笔记
- 🏷️ **智能标签** - 自动分类与手动标记
- 🔍 **全文搜索** - 快速检索任意内容
- ✅ **任务管理** - 待办事项与提醒
- 📊 **洞察报告** - 周报、趋势分析、智能建议
- 🤖 **AI 增强** - 自动摘要、关联推荐
- 🤝 **多 Agent 支持** - News Agent、Tech-Doc Agent 等

## 技术栈

- **后端**: Go 1.24+
- **数据库**: PostgreSQL 16 (生产) / SQLite (开发)
- **前端**: 原生 HTML/CSS/JS (Phase 1)
- **AI**: OpenAI / Anthropic / Local LLM

## 快速开始

### 使用 PostgreSQL（推荐）

```bash
# 1. 启动 PostgreSQL
docker compose up -d postgres

# 2. 验证数据库
./scripts/verify-db.sh

# 3. 复制配置文件
cp config.postgres.yaml ~/.insight-hub/config.yaml

# 4. 编译运行
make build
make run
```

### 使用 SQLite（开发用）

```bash
# 使用默认配置
make run
```

### 数据库管理（可选）

```bash
# 启动 pgAdmin Web UI
docker compose --profile admin up -d pgadmin

# 访问 http://localhost:5050
# 邮箱: admin@insight-hub.local
# 密码: admin
```

## 项目结构

```
insight-hub/
├── cmd/insight-hub/      # 主入口
├── internal/
│   ├── api/              # HTTP handlers
│   ├── config/           # 配置管理
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层
│   └── service/          # 业务逻辑层
├── deploy/
│   └── postgres/         # PostgreSQL 部署文件
├── scripts/              # 工具脚本
├── web/                  # 前端文件
├── docker-compose.yaml   # Docker 编排
├── config.postgres.yaml  # PostgreSQL 配置示例
├── config.example.yaml   # SQLite 配置示例
├── Makefile
└── README.md
```

## 数据库连接

| 参数 | 值 |
|------|-----|
| Host | localhost |
| Port | 5433 |
| Database | insight_hub |
| User | insight |
| Password | insight_local_dev |

## 开发

```bash
# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint
```

## License

MIT
