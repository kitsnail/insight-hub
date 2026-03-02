package model

import "time"

// ============================================
// v3 数据模型 - 个人数据中枢
// ============================================

// ItemStatus 内容状态
type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "active"
	ItemStatusArchived ItemStatus = "archived"
	ItemStatusDeleted  ItemStatus = "deleted"
)

// Item v3 核心模型
// 开放设计：type 和 source_system 都是开放字符串
type Item struct {
	// 主键
	ID string `json:"id"`

	// 数据分类（开放）
	Type     string  `json:"type"`               // 开放类型，不限制
	Category *string `json:"category,omitempty"` // 可选二级分类

	// 来源追溯
	SourceSystem string  `json:"source_system"`          // 数据来源系统
	SourceAgent  *string `json:"source_agent,omitempty"` // 来源 Agent（可选）
	SourceURL    *string `json:"source_url,omitempty"`   // 原始 URL（如有）

	// 核心内容
	Title   string  `json:"title"`
	Content *string `json:"content,omitempty"`
	Summary *string `json:"summary,omitempty"`

	// 灵活元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// 状态与优先级
	Status   ItemStatus `json:"status"`
	Priority int        `json:"priority"`

	// 时间戳
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// 标签（关联）
	Tags []Tag `json:"tags,omitempty"`
}

// ItemCreate 创建 Item 请求
type ItemCreate struct {
	ID           string                 `json:"id"`   // 必填，建议 {source_system}:{uuid}
	Type         string                 `json:"type"` // 必填
	Category     *string                `json:"category,omitempty"`
	SourceSystem string                 `json:"source_system"` // 必填
	SourceAgent  *string                `json:"source_agent,omitempty"`
	SourceURL    *string                `json:"source_url,omitempty"`
	Title        string                 `json:"title"` // 必填
	Content      *string                `json:"content,omitempty"`
	Summary      *string                `json:"summary,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Status       ItemStatus             `json:"status,omitempty"`   // 默认 active
	Priority     int                    `json:"priority,omitempty"` // 默认 0
	OccurredAt   *time.Time             `json:"occurred_at,omitempty"`
	Tags         []string               `json:"tags,omitempty"` // 标签名称列表
}

// ItemUpdate 更新 Item 请求
type ItemUpdate struct {
	Type       *string                `json:"type,omitempty"`
	Category   *string                `json:"category,omitempty"`
	Title      *string                `json:"title,omitempty"`
	Content    *string                `json:"content,omitempty"`
	Summary    *string                `json:"summary,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Status     *ItemStatus            `json:"status,omitempty"`
	Priority   *int                   `json:"priority,omitempty"`
	OccurredAt *time.Time             `json:"occurred_at,omitempty"`
	Tags       []string               `json:"tags,omitempty"` // 替换标签
}

// ItemQuery 查询参数
type ItemQuery struct {
	// 基本过滤
	Type         string `form:"type"`
	Category     string `form:"category"`
	SourceSystem string `form:"source_system"`
	SourceAgent  string `form:"source_agent"`
	Status       string `form:"status"` // 默认 active

	// 时间范围
	DateFrom string `form:"date_from"` // YYYY-MM-DD 或 ISO8601
	DateTo   string `form:"date_to"`

	// 全文搜索
	Search string `form:"search"`

	// 元数据过滤
	Tags string `form:"tags"` // 逗号分隔

	// 分页
	Limit  int `form:"limit" binding:"max=1000"` // 默认 100
	Offset int `form:"offset" binding:"min=0"`

	// 排序
	SortBy   string `form:"sort_by"`   // created_at, updated_at, occurred_at, priority
	SortDesc bool   `form:"sort_desc"` // 默认 true
}

// ItemListResponse 列表响应
type ItemListResponse struct {
	Items   []Item `json:"items"`
	Total   int    `json:"total"`
	Limit   int    `json:"limit"`
	Offset  int    `json:"offset"`
	HasMore bool   `json:"has_more"`
}

// DedupCheckRequest 去重检查请求
type DedupCheckRequest struct {
	SourceSystem string `json:"source_system"`          // 必填
	SourceAgent  string `json:"source_agent,omitempty"` // 可选
	ExternalID   string `json:"external_id,omitempty"`  // 可选，外部系统的唯一标识
	Title        string `json:"title,omitempty"`        // 可选，标题匹配
}

// DedupCheckResponse 去重检查响应
type DedupCheckResponse struct {
	Exists      bool   `json:"exists"`
	ItemID      string `json:"item_id,omitempty"`
	Title       string `json:"title,omitempty"`
	OccurredAt  string `json:"occurred_at,omitempty"`
	DedupReason string `json:"dedup_reason,omitempty"` // external_id, title, content_hash
}

// BatchCreateRequest 批量创建请求
type BatchCreateRequest struct {
	Items []ItemCreate `json:"items"`
}

// BatchCreateResponse 批量创建响应
type BatchCreateResponse struct {
	Created     []Item             `json:"created"`
	Failed      []BatchCreateError `json:"failed"`
	Total       int                `json:"total"`
	Succeeded   int                `json:"succeeded"`
	FailedCount int                `json:"failed_count"`
}

// BatchCreateError 批量创建错误
type BatchCreateError struct {
	Index int    `json:"index"` // 在请求数组中的索引
	ID    string `json:"id"`    // item ID
	Error string `json:"error"` // 错误信息
}

// StatsResponse 统计响应
type StatsResponse struct {
	TotalItems int            `json:"total_items"`
	ByType     map[string]int `json:"by_type"`
	BySource   map[string]int `json:"by_source"`
	ByStatus   map[string]int `json:"by_status"`
	ByTag      map[string]int `json:"by_tag,omitempty"`
	Recent     []Item         `json:"recent,omitempty"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []string `json:"ids"`
}

// BatchDeleteResponse 批量删除响应
type BatchDeleteResponse struct {
	Total     int                `json:"total"`
	Succeeded int                `json:"succeeded"`
	Failed    []BatchDeleteError `json:"failed"`
}

// BatchDeleteError 批量删除错误
type BatchDeleteError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// BatchUpdateStatusRequest 批量更新状态请求
type BatchUpdateStatusRequest struct {
	IDs    []string   `json:"ids"`
	Status ItemStatus `json:"status"`
}

// BatchUpdateStatusResponse 批量更新状态响应
type BatchUpdateStatusResponse struct {
	Total     int                      `json:"total"`
	Succeeded int                      `json:"succeeded"`
	Failed    []BatchUpdateStatusError `json:"failed"`
}

// BatchUpdateStatusError 批量更新状态错误
type BatchUpdateStatusError struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// SourceSystem 来源系统
type SourceSystem struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ============================================
// 其他模型
// ============================================

// Tag 标签模型
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  *string   `json:"category,omitempty"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority 任务优先级
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// Task 任务模型
type Task struct {
	ID          string       `json:"id"`
	ItemID      string       `json:"item_id,omitempty"`
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	DueDate     *time.Time   `json:"due_date,omitempty"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Recurrence  string       `json:"recurrence,omitempty"`
}

// Source 数据源模型
type Source struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Type       string     `json:"type"`
	Config     string     `json:"config,omitempty"` // JSON
	Enabled    bool       `json:"enabled"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ProcessingAction 处理动作
type ProcessingAction string

const (
	ProcessingActionSummarize ProcessingAction = "summarize"
	ProcessingActionTag       ProcessingAction = "tag"
	ProcessingActionRelate    ProcessingAction = "relate"
	ProcessingActionInsight   ProcessingAction = "insight"
)

// ProcessingStatus 处理状态
type ProcessingStatus string

const (
	ProcessingStatusPending    ProcessingStatus = "pending"
	ProcessingStatusProcessing ProcessingStatus = "processing"
	ProcessingStatusDone       ProcessingStatus = "done"
	ProcessingStatusFailed     ProcessingStatus = "failed"
)

// ProcessingQueue AI 处理队列
type ProcessingQueue struct {
	ID          int64            `json:"id"`
	ItemID      string           `json:"item_id"`
	Action      ProcessingAction `json:"action"`
	Status      ProcessingStatus `json:"status"`
	Error       string           `json:"error,omitempty"`
	RetryCount  int              `json:"retry_count"`
	CreatedAt   time.Time        `json:"created_at"`
	ProcessedAt *time.Time       `json:"processed_at,omitempty"`
}

// ============================================
// 预定义常量（可选，用于文档和约定）
// ============================================

// 预定义的来源系统
const (
	SourceOpenclaw   = "openclaw"
	SourcePicoclaw   = "picoclaw"
	SourceOpencode   = "opencode"
	SourceClaudeCode = "claude-code"
	SourceCodex      = "codex"
	SourceManual     = "manual"
	SourceWeb        = "web"
)

// 预定义的类型
const (
	TypeMemory   = "memory"
	TypeNote     = "note"
	TypeNews     = "news"
	TypeDecision = "decision"
	TypeTask     = "task"
	TypeInsight  = "insight"
	TypeTracking = "tracking"
	TypeBrief    = "brief"
	TypeLog      = "log"
	TypeCode     = "code"
	TypeBookmark = "bookmark"
)

// ============================================
// v2 兼容性别名（废弃，仅用于过渡期）
// ============================================

// v2 Agent 常量已废弃，使用 SourceSystem
const (
	AgentNews    = SourceOpenclaw
	AgentTechDoc = SourceOpenclaw
	AgentUnknown = "unknown"
)

// v2 兼容性常量
const (
	ItemTypeBrief    = TypeBrief
	ItemTypeNews     = TypeNews
	ItemTypeTracking = TypeTracking
	ItemTypeInsight  = TypeInsight
	ItemTypeLog      = TypeLog
)

// v2 兼容性元数据结构（已废弃，使用 Metadata JSONB）
type (
	NewsMetadata     map[string]interface{}
	BriefMetadata    map[string]interface{}
	TrackingMetadata map[string]interface{}
	InsightMetadata  map[string]interface{}
	LogMetadata      map[string]interface{}
)

// TrackingHistoryItem 追踪历史项
type TrackingHistoryItem struct {
	Date        string  `json:"date"`
	Summary     string  `json:"summary"`
	Importance  float64 `json:"importance"`
	SourceCount int     `json:"source_count"`
}
