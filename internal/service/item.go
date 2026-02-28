package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/kitsnail/insight-hub/internal/repository"
)

// ItemService Item 服务
type ItemService struct {
	itemRepo *repository.ItemRepository
	tagRepo  *repository.TagRepository
}

// NewItemService 创建 Item 服务
func NewItemService(itemRepo *repository.ItemRepository, tagRepo *repository.TagRepository) *ItemService {
	return &ItemService{
		itemRepo: itemRepo,
		tagRepo:  tagRepo,
	}
}

// CreateItemInput 创建 Item 输入
type CreateItemInput struct {
	Type       model.ItemType `json:"type"`
	Title      string         `json:"title"`
	Content    string         `json:"content"`
	Summary    string         `json:"summary"`
	Source     string         `json:"source"`
	SourceURL  string         `json:"source_url"`
	Tags       []string       `json:"tags"` // 标签名称列表
}

// CreateItem 创建 Item
func (s *ItemService) CreateItem(input CreateItemInput) (*model.Item, error) {
	// 生成 ID
	id := uuid.New().String()

	item := &model.Item{
		ID:      id,
		Type:    input.Type,
		Title:   input.Title,
		Content: input.Content,
		Summary: input.Summary,
		Source:  input.Source,
		SourceURL: input.SourceURL,
	}

	// 处理标签
	if len(input.Tags) > 0 {
		tags := []model.Tag{}
		for _, tagName := range input.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			tag, err := s.tagRepo.GetOrCreateByName(tagName, "")
			if err != nil {
				return nil, fmt.Errorf("failed to create tag %s: %w", tagName, err)
			}
			tags = append(tags, *tag)
		}
		item.Tags = tags
	}

	// 保存 Item
	if err := s.itemRepo.Create(item); err != nil {
		return nil, err
	}

	// 更新全文索引
	if err := s.itemRepo.UpdateFTS(item); err != nil {
		// 索引失败不影响主流程
		fmt.Printf("Warning: failed to update FTS: %v\n", err)
	}

	return item, nil
}

// GetItem 获取 Item
func (s *ItemService) GetItem(id string) (*model.Item, error) {
	return s.itemRepo.GetByID(id)
}

// ListItemsInput 列表查询输入
type ListItemsInput struct {
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Type     model.ItemType `json:"type"`
	Status   model.ItemStatus `json:"status"`
}

// ListItemsOutput 列表查询输出
type ListItemsOutput struct {
	Items    []*model.Item `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// ListItems 列出 Items
func (s *ItemService) ListItems(input ListItemsInput) (*ListItemsOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 20
	}

	offset := (input.Page - 1) * input.PageSize

	items, total, err := s.itemRepo.List(input.PageSize, offset, input.Type, input.Status)
	if err != nil {
		return nil, err
	}

	return &ListItemsOutput{
		Items:    items,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}

// UpdateItemInput 更新 Item 输入
type UpdateItemInput struct {
	Title   *string         `json:"title"`
	Content *string         `json:"content"`
	Summary *string         `json:"summary"`
	Status  *model.ItemStatus `json:"status"`
	Tags    []string        `json:"tags"`
}

// UpdateItem 更新 Item
func (s *ItemService) UpdateItem(id string, input UpdateItemInput) (*model.Item, error) {
	item, err := s.itemRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.Content != nil {
		item.Content = *input.Content
	}
	if input.Summary != nil {
		item.Summary = *input.Summary
	}
	if input.Status != nil {
		item.Status = *input.Status
	}

	// 更新标签
	if input.Tags != nil {
		tagIDs := []int64{}
		for _, tagName := range input.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}
			tag, err := s.tagRepo.GetOrCreateByName(tagName, "")
			if err != nil {
				return nil, fmt.Errorf("failed to create tag %s: %w", tagName, err)
			}
			tagIDs = append(tagIDs, tag.ID)
		}
		if err := s.itemRepo.SetItemTags(id, tagIDs); err != nil {
			return nil, err
		}
	}

	if err := s.itemRepo.Update(item); err != nil {
		return nil, err
	}

	// 更新全文索引
	if err := s.itemRepo.UpdateFTS(item); err != nil {
		fmt.Printf("Warning: failed to update FTS: %v\n", err)
	}

	return item, nil
}

// DeleteItem 删除 Item
func (s *ItemService) DeleteItem(id string) error {
	return s.itemRepo.Delete(id)
}

// SearchItems 搜索 Items
func (s *ItemService) SearchItems(query string, limit int) ([]*model.Item, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.itemRepo.Search(query, limit)
}
