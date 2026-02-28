package repository

import (
	"database/sql"
	"time"

	"github.com/kitsnail/insight-hub/internal/model"
)

// TagRepository Tag 仓储
type TagRepository struct {
	db *Database
}

// NewTagRepository 创建 Tag 仓储
func NewTagRepository(db *Database) *TagRepository {
	return &TagRepository{db: db}
}

// Create 创建 Tag
func (r *TagRepository) Create(tag *model.Tag) error {
	now := time.Now()
	tag.CreatedAt = now

	result, err := r.db.Exec(`
		INSERT INTO tags (name, category, color, created_at)
		VALUES (?, ?, ?, ?)
	`, tag.Name, tag.Category, tag.Color, tag.CreatedAt)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	tag.ID = id

	return nil
}

// GetByID 根据 ID 获取 Tag
func (r *TagRepository) GetByID(id int64) (*model.Tag, error) {
	tag := &model.Tag{}
	var category, color sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, category, color, created_at
		FROM tags WHERE id = ?
	`, id).Scan(&tag.ID, &tag.Name, &category, &color, &tag.CreatedAt)
	if err != nil {
		return nil, err
	}

	tag.Category = category.String
	tag.Color = color.String
	return tag, nil
}

// GetByName 根据名称获取 Tag
func (r *TagRepository) GetByName(name string) (*model.Tag, error) {
	tag := &model.Tag{}
	var category, color sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, category, color, created_at
		FROM tags WHERE name = ?
	`, name).Scan(&tag.ID, &tag.Name, &category, &color, &tag.CreatedAt)
	if err != nil {
		return nil, err
	}

	tag.Category = category.String
	tag.Color = color.String
	return tag, nil
}

// List 列出所有 Tags
func (r *TagRepository) List(category string) ([]model.Tag, error) {
	query := "SELECT id, name, category, color, created_at FROM tags"
	args := []interface{}{}

	if category != "" {
		query += " WHERE category = ?"
		args = append(args, category)
	}

	query += " ORDER BY name"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []model.Tag{}
	for rows.Next() {
		var tag model.Tag
		var cat, color sql.NullString
		err := rows.Scan(&tag.ID, &tag.Name, &cat, &color, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tag.Category = cat.String
		tag.Color = color.String
		tags = append(tags, tag)
	}

	return tags, nil
}

// Update 更新 Tag
func (r *TagRepository) Update(tag *model.Tag) error {
	_, err := r.db.Exec(`
		UPDATE tags SET name = ?, category = ?, color = ?
		WHERE id = ?
	`, tag.Name, tag.Category, tag.Color, tag.ID)
	return err
}

// Delete 删除 Tag
func (r *TagRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}

// GetOrCreateByName 根据名称获取或创建 Tag
func (r *TagRepository) GetOrCreateByName(name string, category string) (*model.Tag, error) {
	tag, err := r.GetByName(name)
	if err == nil {
		return tag, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// 创建新标签
	tag = &model.Tag{
		Name:     name,
		Category: category,
	}
	if err := r.Create(tag); err != nil {
		return nil, err
	}

	return tag, nil
}

// GetItemCounts 获取每个标签关联的 Item 数量
func (r *TagRepository) GetItemCounts() (map[int64]int, error) {
	rows, err := r.db.Query(`
		SELECT tag_id, COUNT(*) as count
		FROM item_tags
		GROUP BY tag_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var tagID int64
		var count int
		if err := rows.Scan(&tagID, &count); err != nil {
			return nil, err
		}
		counts[tagID] = count
	}

	return counts, nil
}
