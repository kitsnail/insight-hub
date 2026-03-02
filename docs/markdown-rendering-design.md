# Markdown 渲染设计方案

> P0 优先级 | 基础能力增强

---

## 1. 需求分析

### 1.1 问题陈述

当前 Insight Hub 的内容展示仅支持纯文本，无法渲染 Markdown 格式。这导致：

| 问题 | 影响 |
|------|------|
| 标题层级不清晰 | 长文档难以阅读 |
| 代码无高亮 | 技术笔记可读性差 |
| 列表无格式 | 结构化内容混乱 |
| 链接不可点击 | 无法跳转外部资源 |
| 表格无样式 | 数据对比困难 |

### 1.2 使用场景

| 场景 | 内容类型 | Markdown 特性 |
|------|---------|--------------|
| 技术笔记 | note | 代码块、标题、列表 |
| 决策记录 | decision | 表格、引用、标题 |
| 追踪话题 | tracking | 链接、列表 |
| 洞察总结 | insight | 全部特性 |
| 简报 | brief | 表格、列表、链接 |

### 1.3 非目标

- ❌ Markdown 编辑器（暂不考虑，使用外部编辑器）
- ❌ 实时预览
- ❌ 服务端渲染

---

## 2. 技术调研

### 2.1 主流方案对比

| 方案 | 周下载量 | 体积 (gzip) | 安全性 | 扩展性 | 适用场景 |
|------|---------|-------------|--------|--------|---------|
| **marked** | ~8M | ~8KB | 中等（历史问题多） | 弱 | 简单渲染 |
| **markdown-it** | ~6M | ~15KB | 高（安全默认） | 强（插件生态） | 通用场景 ✅ |
| **remark (unified)** | ~4M | ~20KB+ | 高 | 最强（AST 中间层） | 复杂处理 |

### 2.2 关键发现

#### 2.2.1 "Don't use marked" 争议

Tom Macwright（Mapbox 联合创始人）在 2024 年发文明确建议：
> "marked is really popular. It used to be the best option. But there are better options, use them!"

**他推荐：**
- **remark** — 最可扩展，插件生态最强
- **markdown-it** — 安全默认，CommonMark 兼容

#### 2.2.2 安全性问题

Markdown 渲染是 XSS 重灾区：
- 2024 年仍有多个 CVE（VNote RCE、markdown-to-jsx XSS）
- **所有解析器都需要配合 XSS 防护**（DOMPurify 是标准做法）
- CSP (Content Security Policy) 是额外防护层

#### 2.2.3 性能对比

根据 2024 基准测试：
```
markdown-it > marked > remark（速度排序）
```

但差异在毫秒级，对用户体验影响可忽略。

---

## 3. 技术方案

### 3.1 方案选择

**推荐方案：markdown-it + highlight.js + DOMPurify**

#### 选择理由

| 维度 | marked | markdown-it ✅ | remark |
|------|--------|---------------|--------|
| **安全性** | 中等（历史漏洞多） | 高（安全默认） | 高 |
| **扩展性** | 弱（无插件系统） | 强（丰富插件生态） | 最强 |
| **体积** | ~8KB | ~15KB (+7KB) | ~20KB+ |
| **学习曲线** | 低 | 低 | 中高 |
| **社区活跃度** | 高 | 高 | 高 |
| **未来扩展** | 需手动集成 | 插件支持 | 插件支持 |

**markdown-it 优势：**
1. **安全性更好** — 安全默认，历史漏洞少
2. **扩展性更强** — 插件生态支持未来需求（数学公式、流程图）
3. **体积可接受** — +7KB 换来更好的架构
4. **社区活跃** — 广泛使用，问题容易找到解决方案

#### remark 不选择的原因

| 问题 | 说明 |
|------|------|
| 过度设计 | 个人项目用不到 AST 中间层 |
| 体积大 | 基础包已是 markdown-it 的 1.5 倍 |
| 学习曲线 | unified 生态概念多，维护成本高 |

### 3.2 插件生态

markdown-it 支持丰富的插件，满足未来扩展需求：

| 功能 | 插件 | 体积 | 优先级 |
|------|------|------|--------|
| 代码高亮 | highlight.js | ~15KB | P0 |
| XSS 防护 | DOMPurify | ~6KB | P0 |
| 表情符号 | markdown-it-emoji | ~2KB | P2 |
| 脚注 | markdown-it-footnote | ~1KB | P2 |
| 数学公式 | @traptitech/markdown-it-katex | ~50KB | P3 |
| 流程图 | markdown-it-mermaid | ~200KB | P3 |
| 高亮标记 | markdown-it-mark | ~1KB | P2 |
| 目录生成 | markdown-it-toc-done-right | ~2KB | P2 |

### 3.3 依赖引入

#### 方案 A：CDN 引入（开发/公网部署）

```html
<!-- Markdown 渲染 -->
<script src="https://cdn.jsdelivr.net/npm/markdown-it@14/dist/markdown-it.min.js"></script>

<!-- 代码高亮 -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/highlight.js@11/styles/github-dark.min.css">
<script src="https://cdn.jsdelivr.net/npm/highlight.js@11"></script>

<!-- XSS 防护 -->
<script src="https://cdn.jsdelivr.net/npm/dompurify@3/dist/purify.min.js"></script>
```

#### 方案 B：本地打包（内网部署）

```html
<!-- 本地依赖 -->
<script src="/static/js/markdown-it.min.js"></script>
<link rel="stylesheet" href="/static/css/github-dark.min.css">
<script src="/static/js/highlight.min.js"></script>
<script src="/static/js/purify.min.js"></script>
```

**本地打包脚本：**
```bash
#!/bin/bash
# scripts/download-markdown-deps.sh

STATIC_DIR="web/static"
mkdir -p $STATIC_DIR/js $STATIC_DIR/css

# markdown-it
curl -o $STATIC_DIR/js/markdown-it.min.js \
  https://cdn.jsdelivr.net/npm/markdown-it@14/dist/markdown-it.min.js

# highlight.js
curl -o $STATIC_DIR/js/highlight.min.js \
  https://cdn.jsdelivr.net/npm/highlight.js@11/build/highlight.min.js
curl -o $STATIC_DIR/css/github-dark.min.css \
  https://cdn.jsdelivr.net/npm/highlight.js@11/styles/github-dark.min.css

# DOMPurify
curl -o $STATIC_DIR/js/purify.min.js \
  https://cdn.jsdelivr.net/npm/dompurify@3/dist/purify.min.js

echo "Markdown dependencies downloaded to $STATIC_DIR"
```

**实施策略：**
- 默认使用 CDN（开发便捷）
- 内网部署时运行下载脚本，切换为本地依赖

---

## 4. 实现设计

### 4.1 核心函数

```javascript
// Markdown 渲染器初始化
function initMarkdownRenderer() {
  // 创建 markdown-it 实例
  window.md = window.markdownit({
    html: false,        // 禁止 HTML 标签（安全考虑）
    xhtmlOut: false,    // 不使用 XHTML 自闭合标签
    breaks: true,       // 支持 GFM 换行
    linkify: true,      // 自动转换 URL 为链接
    typographer: true,  // 启用排版优化
    highlight: function(code, lang) {
      // 代码高亮
      if (lang && hljs.getLanguage(lang)) {
        try {
          return hljs.highlight(code, { language: lang }).value;
        } catch (e) {
          console.warn('Highlight error:', e);
        }
      }
      // 自动检测语言
      return hljs.highlightAuto(code).value;
    }
  });
}

// 渲染 Markdown 内容（带 XSS 防护）
function renderMarkdown(content) {
  if (!content || content.trim() === '') {
    return '<span class="text-gray-500">暂无内容</span>';
  }
  
  // 限制长度（防止性能问题）
  const MAX_LENGTH = 100000; // 100KB
  if (content.length > MAX_LENGTH) {
    content = content.substring(0, MAX_LENGTH) + '\n\n... (内容过长，已截断)';
  }
  
  // 渲染 + XSS 净化
  const rawHtml = window.md.render(content);
  return DOMPurify.sanitize(rawHtml, {
    // 允许的标签（白名单）
    ALLOWED_TAGS: [
      'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
      'p', 'br', 'hr',
      'ul', 'ol', 'li',
      'blockquote', 'pre', 'code',
      'strong', 'em', 'del', 'mark',
      'a', 'img',
      'table', 'thead', 'tbody', 'tr', 'th', 'td',
      'span', 'div'
    ],
    // 允许的属性
    ALLOWED_ATTR: [
      'href', 'src', 'alt', 'title',
      'class', 'id',
      'target', 'rel'
    ],
    // 允许 data-* 属性（代码高亮需要）
    ALLOW_DATA_ATTR: true
  });
}
```

### 4.2 样式设计

```css
/* Markdown 内容样式 */
.markdown-body {
  color: var(--text-primary);
  line-height: 1.7;
  font-size: 0.95em;
}

/* 标题 */
.markdown-body h1, .markdown-body h2, .markdown-body h3,
.markdown-body h4, .markdown-body h5, .markdown-body h6 {
  margin-top: 1.5em;
  margin-bottom: 0.5em;
  font-weight: 600;
  color: var(--text-primary);
}

.markdown-body h1 { font-size: 1.5em; border-bottom: 1px solid var(--glass-border); padding-bottom: 0.3em; }
.markdown-body h2 { font-size: 1.3em; border-bottom: 1px solid var(--glass-border); padding-bottom: 0.3em; }
.markdown-body h3 { font-size: 1.15em; }
.markdown-body h4 { font-size: 1.05em; }

/* 段落 */
.markdown-body p {
  margin: 1em 0;
}

/* 行内代码 */
.markdown-body code {
  background: var(--bg-elevated);
  color: var(--accent-pink);
  padding: 0.2em 0.4em;
  border-radius: 4px;
  font-family: 'Fira Code', 'SF Mono', monospace;
  font-size: 0.9em;
}

/* 代码块 */
.markdown-body pre {
  background: var(--bg-elevated);
  border: 1px solid var(--glass-border);
  padding: 1em;
  border-radius: 8px;
  overflow-x: auto;
  margin: 1em 0;
}

.markdown-body pre code {
  background: transparent;
  color: var(--text-primary);
  padding: 0;
  font-size: 0.85em;
  line-height: 1.5;
}

/* 引用 */
.markdown-body blockquote {
  border-left: 4px solid var(--accent-blue);
  padding: 0.5em 1em;
  margin: 1em 0;
  color: var(--text-secondary);
  background: var(--bg-elevated);
  border-radius: 0 8px 8px 0;
}

/* 列表 */
.markdown-body ul, .markdown-body ol {
  padding-left: 2em;
  margin: 1em 0;
}

.markdown-body li {
  margin: 0.3em 0;
}

/* 表格 */
.markdown-body table {
  width: 100%;
  border-collapse: collapse;
  margin: 1em 0;
  display: block;
  overflow-x: auto;
}

.markdown-body th, .markdown-body td {
  border: 1px solid var(--glass-border);
  padding: 0.5em 1em;
  text-align: left;
}

.markdown-body th {
  background: var(--bg-elevated);
  font-weight: 600;
}

/* 链接 */
.markdown-body a {
  color: var(--accent-cyan);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

.markdown-body a:hover {
  border-bottom-color: var(--accent-cyan);
}

/* 图片 */
.markdown-body img {
  max-width: 100%;
  border-radius: 8px;
  margin: 1em 0;
}

/* 分割线 */
.markdown-body hr {
  border: none;
  border-top: 1px solid var(--glass-border);
  margin: 2em 0;
}

/* 删除线 */
.markdown-body del {
  color: var(--text-secondary);
}

/* 高亮标记 */
.markdown-body mark {
  background: var(--accent-yellow);
  color: var(--bg-primary);
  padding: 0.1em 0.3em;
  border-radius: 2px;
}

/* 任务列表 */
.markdown-body input[type="checkbox"] {
  margin-right: 0.5em;
}
```

### 4.3 应用位置

| 位置 | 当前实现 | 改造后 |
|------|---------|--------|
| 卡片内容预览 | `line-clamp-2` 纯文本 | 保持纯文本（预览不需要渲染） |
| 详情模态框 - 内容 | `whitespace-pre-wrap` | `markdown-body` 渲染 |
| 编辑表单 - 内容 | textarea | 保持 textarea（编辑用） |

**判断：** 仅在详情模态框中渲染 Markdown，卡片预览保持纯文本（性能考虑）

---

## 5. 边界情况

### 5.1 按类型控制渲染

```javascript
// 某些类型可能不适合 Markdown 渲染
const MARKDOWN_TYPES = ['note', 'decision', 'insight', 'brief', 'tracking'];

function shouldRenderMarkdown(type) {
  return MARKDOWN_TYPES.includes(type);
}

function renderItemContent(item) {
  if (shouldRenderMarkdown(item.type) && item.content) {
    return `<div class="markdown-body">${renderMarkdown(item.content)}</div>`;
  }
  return `<p class="whitespace-pre-wrap">${escapeHtml(item.content || '')}</p>`;
}

// HTML 转义（用于非 Markdown 内容）
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}
```

### 5.2 代码块复制按钮

```javascript
// 为代码块添加复制按钮
function addCopyButtons() {
  document.querySelectorAll('.markdown-body pre').forEach(pre => {
    const button = document.createElement('button');
    button.className = 'copy-btn absolute top-2 right-2 px-2 py-1 text-xs bg-glass-bg rounded opacity-0 group-hover:opacity-100 transition-opacity';
    button.textContent = '复制';
    
    button.addEventListener('click', async () => {
      const code = pre.querySelector('code').textContent;
      await navigator.clipboard.writeText(code);
      button.textContent = '已复制!';
      setTimeout(() => button.textContent = '复制', 2000);
    });
    
    pre.style.position = 'relative';
    pre.classList.add('group');
    pre.appendChild(button);
  });
}
```

### 5.3 图片懒加载

```javascript
// 图片懒加载（已在渲染后处理）
function lazyLoadImages() {
  document.querySelectorAll('.markdown-body img').forEach(img => {
    img.loading = 'lazy';
    img.addEventListener('error', function() {
      this.src = '/static/img/broken-image.svg';
      this.alt = '图片加载失败';
    });
  });
}
```

---

## 6. 测试用例

### 6.1 基础语法

```markdown
# 标题一

## 标题二

### 标题三

普通段落文本，支持**加粗**、*斜体*、`代码`、~~删除线~~。

[链接文本](https://example.com)

![图片描述](https://example.com/image.png)
```

### 6.2 代码块

````markdown
```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

```python
def hello():
    print("Hello, World!")
```

```bash
echo "Hello, World!"
```
````

### 6.3 列表

```markdown
- 无序列表项 1
- 无序列表项 2
  - 嵌套项

1. 有序列表项 1
2. 有序列表项 2

- [ ] 任务列表（未完成）
- [x] 任务列表（已完成）
```

### 6.4 表格

```markdown
| 列 1 | 列 2 | 列 3 |
|------|------|------|
| 数据 | 数据 | 数据 |
| 数据 | 数据 | 数据 |
```

### 6.5 引用

```markdown
> 这是一段引用文本
> 可以多行
```

### 6.6 安全测试

```markdown
<!-- 应该被过滤 -->
<script>alert('XSS')</script>

<img src=x onerror=alert('XSS')>

<a href="javascript:alert('XSS')">点击</a>

<!-- 应该正常显示 -->
**安全的 Markdown**
```

---

## 7. 性能考虑

### 7.1 渲染时机

- ✅ 懒加载：仅在打开详情模态框时渲染
- ❌ 预渲染：不在列表加载时渲染（性能浪费）

### 7.2 缓存策略

```javascript
// 可选：缓存渲染结果（LRU 策略）
class MarkdownCache {
  constructor(maxSize = 50) {
    this.cache = new Map();
    this.maxSize = maxSize;
  }
  
  get(content) {
    return this.cache.get(content);
  }
  
  set(content, result) {
    if (this.cache.size >= this.maxSize) {
      // 删除最旧的条目
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }
    this.cache.set(content, result);
  }
}

const mdCache = new MarkdownCache(50);

function renderMarkdownCached(content) {
  const cached = mdCache.get(content);
  if (cached) return cached;
  
  const result = renderMarkdown(content);
  mdCache.set(content, result);
  return result;
}
```

### 7.3 体积影响

| 资源 | 大小 (gzip) | 说明 |
|------|------------|------|
| markdown-it | ~15KB | 核心渲染 |
| highlight.js | ~15KB | 代码高亮 |
| DOMPurify | ~6KB | XSS 防护 |
| **总计** | **~36KB** | 可接受 |

**对比原方案（marked）：** +7KB，换来更好的扩展性和安全性

---

## 8. 实施计划

### Phase 1：基础实现（P0）

1. ✅ 引入依赖（markdown-it, highlight.js, DOMPurify）
2. ✅ 添加 Markdown 样式
3. ✅ 修改详情模态框渲染逻辑
4. ✅ 测试基础语法
5. ✅ 安全测试（XSS 防护验证）

**预计工作量：** 2-3 小时

### Phase 2：优化增强（P1）

1. 按类型控制渲染
2. 添加代码块复制按钮
3. 图片懒加载
4. 渲染缓存

**预计工作量：** 1-2 小时

### Phase 3：扩展功能（P2/P3）

1. **P2：** 表情符号、脚注、高亮标记
2. **P3：** 数学公式（KaTeX）
3. **P3：** 流程图（Mermaid）
4. **P2：** 目录导航（TOC）

**按需实施，暂不排期**

---

## 9. 风险与缓解

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|---------|
| XSS 攻击 | 高 | 中 | DOMPurify 净化 + 白名单策略 |
| 超长内容卡顿 | 中 | 低 | 限制长度 + 警告提示 |
| CDN 不可用 | 中 | 低 | 本地打包方案 |
| 样式冲突 | 低 | 中 | 使用 `.markdown-body` 命名空间 |
| 插件兼容性 | 低 | 低 | 仅使用成熟插件 |

---

## 10. 决策记录

| 决策 | 选择 | 理由 |
|------|------|------|
| 渲染库 | **markdown-it** | 安全默认 + 扩展性强 + 社区活跃 |
| 代码高亮 | highlight.js | 语法丰富、主题支持 |
| XSS 防护 | DOMPurify | 专用的 HTML 净化库，行业标准 |
| 渲染位置 | 详情模态框 | 性能考虑，预览保持纯文本 |
| 渲染时机 | 懒加载 | 打开详情时渲染 |
| 部署方式 | CDN + 本地备选 | 开发便捷 + 内网兼容 |

---

## 变更记录

| 日期 | 变更 |
|------|------|
| 2026-03-02 | 初始设计（marked 方案） |
| 2026-03-02 | 方案调整：marked → markdown-it（基于调研结果） |
| 2026-03-02 | 添加安全性分析、扩展性分析、插件生态 |
| 2026-03-02 | 添加内网部署方案（本地打包） |
