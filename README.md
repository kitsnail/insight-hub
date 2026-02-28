# Insight Hub

个人知识管理中枢 - 收集、整理、洞察。

## 功能特性

- 📥 **多源数据收集** - 浏览器书签、聊天记录、代码片段、笔记
- 🏷️ **智能标签** - 自动分类与手动标记
- 🔍 **全文搜索** - 快速检索任意内容
- ✅ **任务管理** - 待办事项与提醒
- 📊 **洞察报告** - 周报、趋势分析、智能建议
- 🤖 **AI 增强** - 自动摘要、关联推荐

## 技术栈

- **后端**: Go 1.21+
- **数据库**: SQLite (WAL mode)
- **前端**: 原生 HTML/CSS/JS (Phase 1)
- **AI**: OpenAI / Anthropic / Local LLM

## 快速开始

### 编译

```bash
make build
```

### 运行

```bash
# 使用默认配置
make run

# 使用自定义配置
./bin/insight-hub -config ~/.insight-hub/config.yaml
```

### 初始化

```bash
make db-init
```

## 配置

复制示例配置文件：

```bash
cp config.example.yaml ~/.insight-hub/config.yaml
```

编辑 `~/.insight-hub/config.yaml` 配置 LLM API Key 等参数。

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
├── web/                  # 前端文件
├── scripts/              # 工具脚本
├── Makefile
└── README.md
```

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
