package repository

import (
	"context"
	"time"

	"github.com/kitsnail/insight-hub/internal/model"
)

// ItemRepositoryV3 Item 仓储接口 v3
type ItemRepositoryV3 interface {
	// CRUD
	Create(ctx context.Context, item *model.ItemCreate) (*model.Item, error)
	GetByID(ctx context.Context, id string) (*model.Item, error)
	Update(ctx context.Context, id string, update *model.ItemUpdate) (*model.Item, error)
	Delete(ctx context.Context, id string) error

	// 查询
	List(ctx context.Context, query *model.ItemQuery) (*model.ItemListResponse, error)
	Search(ctx context.Context, query string, limit int) ([]*model.Item, error)

	// 批量操作
	BatchCreate(ctx context.Context, items []*model.ItemCreate) ([]*model.Item, error)
	BatchDelete(ctx context.Context, ids []string) (*model.BatchDeleteResponse, error)
	BatchUpdateStatus(ctx context.Context, ids []string, status model.ItemStatus) (*model.BatchUpdateStatusResponse, error)

	// 去重
	CheckDuplicate(ctx context.Context, req *model.DedupCheckRequest) (*model.DedupCheckResponse, error)

	// 统计
	Stats(ctx context.Context) (*model.StatsResponse, error)
	CountByType(ctx context.Context) (map[string]int, error)
	CountBySource(ctx context.Context) (map[string]int, error)

	// 标签
	SetItemTags(ctx context.Context, itemID string, tagNames []string) error
	GetItemTags(ctx context.Context, itemID string) ([]model.Tag, error)
}

// TagRepository Tag 仓储接口
type TagRepository interface {
	Create(ctx context.Context, tag *model.Tag) error
	GetByID(ctx context.Context, id int64) (*model.Tag, error)
	GetByName(ctx context.Context, name string) (*model.Tag, error)
	List(ctx context.Context, category string) ([]model.Tag, error)
	Update(ctx context.Context, tag *model.Tag) error
	Delete(ctx context.Context, id int64) error
	GetOrCreateByName(ctx context.Context, name string, category string) (*model.Tag, error)
	GetItemCounts(ctx context.Context) (map[int64]int, error)
}

// TaskRepository Task 仓储接口
type TaskRepository interface {
	Create(ctx context.Context, task *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	List(ctx context.Context, status model.TaskStatus, limit, offset int) ([]*model.Task, int, error)
	Update(ctx context.Context, task *model.Task) error
	Delete(ctx context.Context, id string) error
	GetByItemID(ctx context.Context, itemID string) ([]*model.Task, error)
}

// DedupRepository 去重仓储接口
type DedupRepository interface {
	Check(ctx context.Context, fingerprint string) (bool, string, error)
	Add(ctx context.Context, fingerprint, itemID, agent string, expiresAt *time.Time) error
	Cleanup(ctx context.Context) (int64, error)
}

// SourceSystemRepository 来源系统仓储接口
type SourceSystemRepository interface {
	Create(ctx context.Context, system *model.SourceSystem) error
	GetByName(ctx context.Context, name string) (*model.SourceSystem, error)
	List(ctx context.Context) ([]model.SourceSystem, error)
}
