package service

import (
	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/kitsnail/insight-hub/internal/repository"
)

// TagService Tag 服务
type TagService struct {
	tagRepo *repository.TagRepository
}

// NewTagService 创建 Tag 服务
func NewTagService(tagRepo *repository.TagRepository) *TagService {
	return &TagService{tagRepo: tagRepo}
}

// CreateTagInput 创建 Tag 输入
type CreateTagInput struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Color    string `json:"color"`
}

// CreateTag 创建 Tag
func (s *TagService) CreateTag(input CreateTagInput) (*model.Tag, error) {
	tag := &model.Tag{
		Name:     input.Name,
		Category: input.Category,
		Color:    input.Color,
	}
	if err := s.tagRepo.Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

// GetTag 获取 Tag
func (s *TagService) GetTag(id int64) (*model.Tag, error) {
	return s.tagRepo.GetByID(id)
}

// ListTags 列出 Tags
func (s *TagService) ListTags(category string) ([]model.Tag, error) {
	return s.tagRepo.List(category)
}

// UpdateTagInput 更新 Tag 输入
type UpdateTagInput struct {
	Name     *string `json:"name"`
	Category *string `json:"category"`
	Color    *string `json:"color"`
}

// UpdateTag 更新 Tag
func (s *TagService) UpdateTag(id int64, input UpdateTagInput) (*model.Tag, error) {
	tag, err := s.tagRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		tag.Name = *input.Name
	}
	if input.Category != nil {
		tag.Category = *input.Category
	}
	if input.Color != nil {
		tag.Color = *input.Color
	}

	if err := s.tagRepo.Update(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

// DeleteTag 删除 Tag
func (s *TagService) DeleteTag(id int64) error {
	return s.tagRepo.Delete(id)
}

// TagWithCount 带数量的 Tag
type TagWithCount struct {
	model.Tag
	ItemCount int `json:"item_count"`
}

// ListTagsWithCount 列出带数量的 Tags
func (s *TagService) ListTagsWithCount(category string) ([]TagWithCount, error) {
	tags, err := s.tagRepo.List(category)
	if err != nil {
		return nil, err
	}

	counts, err := s.tagRepo.GetItemCounts()
	if err != nil {
		return nil, err
	}

	result := make([]TagWithCount, len(tags))
	for i, tag := range tags {
		result[i] = TagWithCount{
			Tag:       tag,
			ItemCount: counts[tag.ID],
		}
	}

	return result, nil
}
