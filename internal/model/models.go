package model

import "time"

// ItemType 内容类型
type ItemType string

const (
	ItemTypeBookmark ItemType = "bookmark"
	ItemTypeNote     ItemType = "note"
	ItemTypeChat     ItemType = "chat"
	ItemTypeCode     ItemType = "code"
	ItemTypeTask     ItemType = "task"
)

// ItemStatus 内容状态
type ItemStatus string

const (
	ItemStatusActive   ItemStatus = "active"
	ItemStatusArchived ItemStatus = "archived"
	ItemStatusDeleted  ItemStatus = "deleted"
)

// Item 核心内容模型
type Item struct {
	ID             string     `json:"id"`
	Type           ItemType   `json:"type"`
	Title          string     `json:"title,omitempty"`
	Content        string     `json:"content,omitempty"`
	Summary        string     `json:"summary,omitempty"`
	Source         string     `json:"source"`
	SourceURL      string     `json:"source_url,omitempty"`
	SourceMetadata string     `json:"source_metadata,omitempty"` // JSON
	OccurredAt     *time.Time `json:"occurred_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Status         ItemStatus `json:"status"`
	Tags           []Tag      `json:"tags,omitempty"` // 关联标签
}

// Tag 标签模型
type Tag struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Category  string    `json:"category,omitempty"`
	Color     string    `json:"color,omitempty"`
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

// InsightType 洞察类型
type InsightType string

const (
	InsightTypeWeeklyReport InsightType = "weekly_report"
	InsightTypeTrend        InsightType = "trend"
	InsightTypeReminder     InsightType = "reminder"
	InsightTypeSuggestion   InsightType = "suggestion"
)

// Insight 洞察模型
type Insight struct {
	ID           string      `json:"id"`
	Type         InsightType `json:"type"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	PeriodStart  *time.Time  `json:"period_start,omitempty"`
	PeriodEnd    *time.Time  `json:"period_end,omitempty"`
	ItemIDs      string      `json:"item_ids,omitempty"` // JSON array
	GeneratedBy  string      `json:"generated_by"`
	CreatedAt    time.Time   `json:"created_at"`
	Status       string      `json:"status"`
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
