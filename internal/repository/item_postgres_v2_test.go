package repository

import (
	"testing"
	"time"

	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/stretchr/testify/assert"
)

// TestItemModelV2 测试 v2 数据模型
func TestItemModelV2(t *testing.T) {
	// 测试类型常量
	assert.Equal(t, model.ItemType("brief"), model.ItemTypeBrief)
	assert.Equal(t, model.ItemType("news"), model.ItemTypeNews)
	assert.Equal(t, model.ItemType("tracking"), model.ItemTypeTracking)
	assert.Equal(t, model.ItemType("insight"), model.ItemTypeInsight)
	assert.Equal(t, model.ItemType("log"), model.ItemTypeLog)

	// 测试 Agent 常量
	assert.Equal(t, "news-agent", model.AgentNews)
	assert.Equal(t, "tech-doc-agent", model.AgentTechDoc)
	assert.Equal(t, "unknown", model.AgentUnknown)
}

// TestNewsMetadata 测试新闻元数据
func TestNewsMetadata(t *testing.T) {
	metadata := model.NewsMetadata{
		Author:      "张三",
		PublishedAt: "2026-03-01T10:00:00Z",
		Category:    "AI",
		Keywords:    []string{"GPT", "LLM"},
		Summary:     "测试摘要",
		Sentiment:   "positive",
		Importance:  0.8,
	}

	assert.Equal(t, "张三", metadata.Author)
	assert.Equal(t, []string{"GPT", "LLM"}, metadata.Keywords)
	assert.Equal(t, 0.8, metadata.Importance)
}

// TestTrackingMetadata 测试追踪元数据
func TestTrackingMetadata(t *testing.T) {
	metadata := model.TrackingMetadata{
		TrackID:       "track-001",
		TrackType:     "trend",
		Topic:         "AI 进展",
		Status:        "active",
		LastCheck:     "2026-03-01T00:00:00Z",
		CheckInterval: "24h",
		History: []model.TrackingHistoryItem{
			{
				Date:        "2026-03-01",
				Summary:     "OpenAI 发布 GPT-5",
				Importance:  0.9,
				SourceCount: 5,
			},
		},
	}

	assert.Equal(t, "track-001", metadata.TrackID)
	assert.Equal(t, "trend", metadata.TrackType)
	assert.Len(t, metadata.History, 1)
}

// TestQueryFilter 测试查询过滤器
func TestQueryFilter(t *testing.T) {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)

	filter := QueryFilter{
		Type:     "news",
		Agent:    "news-agent",
		DateFrom: &from,
		DateTo:   &to,
		Search:   "GPT",
		Limit:    50,
		Offset:   0,
	}

	assert.Equal(t, "news", filter.Type)
	assert.Equal(t, "news-agent", filter.Agent)
	assert.Equal(t, "GPT", filter.Search)
	assert.Equal(t, 50, filter.Limit)
}

// TestItemWithMetadata 测试 Item 带元数据
func TestItemWithMetadata(t *testing.T) {
	now := time.Now()
	occurredAt := now.Add(-1 * time.Hour)

	item := &model.Item{
		ID:      "test-001",
		Type:    model.ItemTypeNews,
		Agent:   model.AgentNews,
		Title:   "测试新闻",
		Content: "这是测试内容",
		Source:  "test-source",
		SourceMetadata: map[string]interface{}{
			"author":      "张三",
			"category":    "AI",
			"importance":  0.8,
		},
		OccurredAt: &occurredAt,
		Status:     model.ItemStatusActive,
	}

	assert.Equal(t, "test-001", item.ID)
	assert.Equal(t, model.ItemTypeNews, item.Type)
	assert.Equal(t, model.AgentNews, item.Agent)
	assert.NotNil(t, item.SourceMetadata)
	assert.Equal(t, "张三", item.SourceMetadata["author"])
}
