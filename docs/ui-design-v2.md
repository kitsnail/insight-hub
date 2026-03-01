# Insight Hub UI Design v2

## 设计目标

将 Insight Hub 从简单的笔记应用界面，升级为现代化的**个人数据中枢 Dashboard**。

## 设计系统

### 风格定位
- **主风格**: Bento Box Grid (Apple-style modular cards)
- **主题**: Dark Mode (OLED 优化)
- **元素**: Glassmorphism (卡片、模态框)
- **氛围**: 技术、数据驱动、现代

### 颜色方案

```css
/* Dark Mode */
--bg-primary: #0A0A0A;      /* 深黑背景 */
--bg-secondary: #141414;    /* 卡片背景 */
--bg-elevated: #1A1A1A;     /* 悬浮元素 */
--text-primary: #F5F5F5;    /* 主文本 */
--text-secondary: #A3A3A3;  /* 次要文本 */
--text-muted: #737373;      /* 弱化文本 */

/* Accent Colors */
--accent-blue: #3B82F6;     /* 主强调色 */
--accent-cyan: #06B6D4;     /* 信息/链接 */
--accent-green: #10B981;    /* 成功/active */
--accent-orange: #F97316;   /* 警告/CTA */
--accent-red: #EF4444;      /* 错误/deleted */

/* Glass Effect */
--glass-bg: rgba(255, 255, 255, 0.05);
--glass-border: rgba(255, 255, 255, 0.1);
--glass-blur: 12px;
```

### 字体

```css
/* Primary: Fira Sans (UI) */
font-family: 'Fira Sans', -apple-system, sans-serif;

/* Monospace: Fira Code (数据、代码) */
font-family: 'Fira Code', 'SF Mono', monospace;

/* 字号系统 */
--text-xs: 0.75rem;    /* 12px */
--text-sm: 0.875rem;   /* 14px */
--text-base: 1rem;     /* 16px */
--text-lg: 1.125rem;   /* 18px */
--text-xl: 1.25rem;    /* 20px */
--text-2xl: 1.5rem;    /* 24px */
--text-3xl: 1.875rem;  /* 30px */
```

### 间距与圆角

```css
/* 间距 */
--space-1: 0.25rem;   /* 4px */
--space-2: 0.5rem;    /* 8px */
--space-3: 0.75rem;   /* 12px */
--space-4: 1rem;      /* 16px */
--space-6: 1.5rem;    /* 24px */
--space-8: 2rem;      /* 32px */

/* 圆角 */
--radius-sm: 0.375rem;  /* 6px */
--radius-md: 0.5rem;    /* 8px */
--radius-lg: 0.75rem;   /* 12px */
--radius-xl: 1rem;      /* 16px */
--radius-2xl: 1.5rem;   /* 24px */
```

## 页面布局

### 1. 整体结构

```
┌─────────────────────────────────────────────────────────────────┐
│  [Navbar] - Logo + 搜索 + 全局操作                              │
├─────────────────────────────────────────────────────────────────┤
│  [Stats Bar] - 统计卡片 (Bento Grid)                            │
├───────────────┬─────────────────────────────────────────────────┤
│  [Sidebar]    │  [Main Content]                                 │
│  - 类型筛选    │  - 工具栏 (排序、视图切换)                       │
│  - 来源筛选    │  - 内容网格/列表                                │
│  - 状态筛选    │  - 分页/无限滚动                                │
│  - 标签云     │                                                 │
└───────────────┴─────────────────────────────────────────────────┘
```

### 2. Navbar

```
┌─────────────────────────────────────────────────────────────────┐
│  ⛵ Insight Hub    [🔍 搜索...]    [Dark/Light] [+ 新建]        │
│  ─────────────────────────────────────────────────────────────  │
│  个人数据中枢                                                    │
└─────────────────────────────────────────────────────────────────┘

特点:
- 浮动样式 (top-4 left-4 right-4)
- Glassmorphism 背景
- 搜索框全局可用
```

### 3. Stats Bar (Bento Grid)

```
┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────────────┐
│ 📊 总数    │ │ 📰 新闻    │ │ 🎯 追踪    │ │ 📈 来源分布        │
│    10      │ │     4      │ │     6      │ │ picoclaw: 10       │
│  +2 今日   │ │  +1 今日   │ │  active: 6 │ │ ━━━━━━━━━━ 100%   │
└────────────┘ └────────────┘ └────────────┘ └────────────────────┘

特点:
- 不对称网格布局
- 卡片大小根据重要性变化
- 数据实时更新
- Hover 效果 (scale 1.02)
```

### 4. Sidebar

```
┌─────────────────┐
│ 类型            │
│ ─────────────── │
│ ○ 全部      10  │
│ ● tracking   6  │
│ ○ news       4  │
│                 │
│ 来源系统        │
│ ─────────────── │
│ ○ picoclaw   10 │
│                 │
│ 状态            │
│ ─────────────── │
│ ● active    10  │
│ ○ deleted   18  │
│                 │
│ 标签            │
│ ─────────────── │
│ AI · 关税 · ... │
└─────────────────┘

特点:
- 动态生成 (从 /api/v1/stats 获取)
- 可折叠分组
- 计数实时更新
```

### 5. 内容卡片

#### Tracking 类型卡片

```
┌─────────────────────────────────────────────────────────────────┐
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ 🎯 特朗普关税政策                          T006 · ● active │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│ 最新进展: 最高法院裁定关税手段违宪 → 特朗普推出全球15%...        │
│                                                                 │
│ ┌─ 关键词 ─────────────────────────────────────────────────┐   │
│ │ 特朗普 · 关税 · Trade Act 122 · 全球关税 · 供应链       │   │
│ └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│ ┌─ 后续关注 ───────────────────────────────────────────────┐   │
│ │ • 政策法律挑战                                          │   │
│ │ • 各国反应                                              │   │
│ │ • 对全球供应链影响                                      │   │
│ └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│ 影响: 对华政策安抚+对全球强硬，暴露战略焦虑                     │
│ ─────────────────────────────────────────────────────────────  │
│ 📅 2026-02-28 · picoclaw/news-agent                            │
└─────────────────────────────────────────────────────────────────┘

特点:
- 类型图标 (SVG, 不用 emoji)
- 状态指示器 (彩色圆点)
- metadata 字段结构化展示
- Hover 效果 (边框发光)
```

#### News 类型卡片

```
┌─────────────────────────────────────────────────────────────────┐
│ 📰 OpenAI 发布 GPT-5: 多模态能力大幅提升                        │
│ ─────────────────────────────────────────────────────────────── │
│ 摘要: OpenAI 今日宣布 GPT-5 正式发布，支持文本、图像、           │
│ 音频、视频的全模态理解和生成...                                 │
│                                                                 │
│ ┌─────────────────────────────────────────────────────────────┐ │
│ │ AI · GPT-5 · OpenAI · 多模态                              │ │
│ └─────────────────────────────────────────────────────────────┘ │
│                                                                 │
│ 💬 positive · ⭐ 重要性: 0.85                                   │
│ ─────────────────────────────────────────────────────────────── │
│ CNBC · 2026-03-01 · picoclaw/news-agent                        │
└─────────────────────────────────────────────────────────────────┘
```

### 6. 模态框 (详情/编辑)

```
┌─────────────────────────────────────────────────────────────────┐
│ [Backdrop Blur Overlay]                                        │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │ 🎯 特朗普关税政策                          [✕] [✎] [🗑]│  │
│   │ ─────────────────────────────────────────────────────── │  │
│   │                                                         │  │
│   │ [Tab: 详情 | 时间线 | 相关内容]                         │  │
│   │                                                         │  │
│   │ 最新进展:                                               │  │
│   │ 最高法院裁定关税手段违宪 → 特朗普推出全球15%关税...     │  │
│   │                                                         │  │
│   │ 关键词: 特朗普, 关税, Trade Act 122...                  │  │
│   │                                                         │  │
│   │ 后续关注:                                               │  │
│   │ • 政策法律挑战                                          │  │
│   │ • 各国反应                                              │  │
│   │                                                         │  │
│   │ 影响:                                                   │  │
│   │ 对华政策安抚+对全球强硬，暴露战略焦虑                   │  │
│   │                                                         │  │
│   │ ─────────────────────────────────────────────────────── │  │
│   │ 创建: 2026-02-28 · 更新: 2026-02-28                     │  │
│   │ 来源: picoclaw / news-agent                             │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

特点:
- Glassmorphism 背景
- 大尺寸 (max-w-3xl)
- Tab 切换不同视图
- 操作按钮在标题栏
```

## 交互设计

### Hover 效果

```css
/* 卡片 Hover */
.item-card {
  transition: all 0.2s ease;
  border: 1px solid transparent;
}
.item-card:hover {
  transform: translateY(-2px);
  border-color: var(--accent-blue);
  box-shadow: 0 0 20px rgba(59, 130, 246, 0.15);
}

/* 按钮 Hover */
.btn-primary:hover {
  background: var(--accent-blue);
  box-shadow: 0 0 15px rgba(59, 130, 246, 0.3);
}
```

### 过渡动画

```css
/* 页面过渡 */
.page-enter { opacity: 0; transform: translateY(10px); }
.page-enter-active { opacity: 1; transform: translateY(0); transition: all 0.3s ease; }

/* 模态框 */
.modal-enter { opacity: 0; transform: scale(0.95); }
.modal-enter-active { opacity: 1; transform: scale(1); transition: all 0.2s ease; }

/* 尊重用户偏好 */
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

### 加载状态

```css
/* Skeleton Loading */
.skeleton {
  background: linear-gradient(90deg, 
    var(--bg-secondary) 25%, 
    var(--bg-elevated) 50%, 
    var(--bg-secondary) 75%
  );
  background-size: 200% 100%;
  animation: skeleton-loading 1.5s infinite;
}

@keyframes skeleton-loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}
```

## 响应式设计

### 断点

```css
/* Mobile First */
/* sm: 640px */
/* md: 768px */
/* lg: 1024px */
/* xl: 1280px */
/* 2xl: 1536px */
```

### 布局变化

```
Mobile (< 768px):
┌─────────────┐
│ Navbar      │
├─────────────┤
│ Stats       │
│ (2x2 grid)  │
├─────────────┤
│ Filter Bar  │
├─────────────┤
│ Cards       │
│ (1 column)  │
└─────────────┘

Tablet (768px - 1024px):
┌─────────────────────┐
│ Navbar              │
├─────────────────────┤
│ Stats (4 columns)   │
├───────┬─────────────┤
│ Side  │ Cards       │
│ bar   │ (2 columns) │
└───────┴─────────────┘

Desktop (> 1024px):
┌─────────────────────────────────────┐
│ Navbar                              │
├─────────────────────────────────────┤
│ Stats (Bento Grid)                  │
├───────┬─────────────────────────────┤
│ Side  │ Cards                       │
│ bar   │ (3 columns)                 │
└───────┴─────────────────────────────┘
```

## API 对接

### 使用 v3 API

```javascript
// Base URL
const API_V1 = '/api/v1';

// 统计数据
GET /api/v1/stats
Response: {
  total: 10,
  by_type: { tracking: 6, news: 4 },
  by_source_system: { picoclaw: 10 },
  by_status: { active: 10, deleted: 18 }
}

// 列表查询
GET /api/v1/items?type=tracking&source_system=picoclaw&status=active&sort=created_at&order=desc&limit=20&offset=0

// 搜索
GET /api/v1/items?q=特朗普

// 详情
GET /api/v1/items/{id}

// 创建
POST /api/v1/items
Body: { type, title, content, source_system, source_agent, metadata, tags }

// 更新
PUT /api/v1/items/{id}

// 删除
DELETE /api/v1/items/{id}
```

### 动态类型列表

```javascript
// 从 stats API 获取类型列表
async function loadTypeFilters() {
  const stats = await fetch('/api/v1/stats').then(r => r.json());
  return Object.entries(stats.by_type).map(([type, count]) => ({
    type,
    count,
    label: getTypeLabel(type)
  }));
}
```

## 图标规范

**禁止使用 Emoji 作为图标**，使用 SVG 图标库：

- **Heroicons**: https://heroicons.com/
- **Lucide**: https://lucide.dev/

### 图标映射

```javascript
const TYPE_ICONS = {
  tracking: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
      d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
  </svg>`,
  news: `<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
      d="M19 20H5a2 2 0 01-2-2V6a2 2 0 012-2h10a2 2 0 012 2v1m2 13a2 2 0 01-2-2V7m2 13a2 2 0 002-2V9a2 2 0 00-2-2h-2m-4-3H9M7 16h6M7 8h6v4H7V8z"/>
  </svg>`,
  // ... 其他类型
};
```

## 实现清单

### Phase 1: 基础框架
- [ ] 创建新的 HTML 文件 (`index-v2.html`)
- [ ] 引入 Tailwind CSS CDN
- [ ] 引入 Google Fonts (Fira Sans, Fira Code)
- [ ] 设置 CSS 变量 (颜色、间距、圆角)
- [ ] 创建基础布局组件

### Phase 2: 核心组件
- [ ] Navbar 组件 (浮动、glass 效果)
- [ ] Stats Bar 组件 (Bento Grid)
- [ ] Sidebar 组件 (筛选器)
- [ ] Item Card 组件 (类型感知)
- [ ] Modal 组件 (详情/编辑)

### Phase 3: API 集成
- [ ] 对接 v3 API
- [ ] 动态类型列表
- [ ] 搜索功能
- [ ] 分页/无限滚动

### Phase 4: 交互优化
- [ ] Hover 效果
- [ ] 过渡动画
- [ ] 加载状态
- [ ] 错误处理

### Phase 5: 响应式
- [ ] Mobile 布局
- [ ] Tablet 布局
- [ ] Desktop 布局
- [ ] 触摸优化

## 文件结构

```
web/
├── static/
│   ├── index.html          # 旧版 (保留)
│   ├── index-v2.html       # 新版入口
│   ├── css/
│   │   └── styles.css      # 自定义样式
│   └── js/
│       ├── app.js          # 主应用逻辑
│       ├── api.js          # API 封装
│       └── components.js   # UI 组件
```

## 预览

完成后的效果应该是：

1. **深色主题** - 护眼、现代、专业
2. **Bento Grid** - 统计数据一目了然
3. **类型感知** - 不同类型数据展示不同字段
4. **Glass 效果** - 卡片、模态框有质感
5. **流畅交互** - Hover、过渡、加载状态完善
6. **响应式** - 各设备体验良好

---

**设计文档版本**: v2.0
**创建日期**: 2026-03-01
**参考规范**: ui-ux-pro-max skill
