package repository

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitsnail/insight-hub/internal/model"
)

// TagRepoPG Tag 仓储 (PostgreSQL 版本)
type TagRepoPG struct {
	pool *pgxpool.Pool
}

// NewTagRepoPG 创建 Tag 仓储
func NewTagRepoPG(pool *pgxpool.Pool) *TagRepoPG {
	return &TagRepoPG{pool: pool}
}

// Create 创建 Tag
func (r *TagRepoPG) Create(ctx context.Context, tag *model.Tag) error {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO tags (name, category, color, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`, tag.Name, tag.Category, tag.Color).Scan(&tag.ID, &tag.CreatedAt)
	return err
}

// GetByID 根据 ID 获取 Tag
func (r *TagRepoPG) GetByID(ctx context.Context, id int64) (*model.Tag, error) {
	tag := &model.Tag{}
	var category, color sql.NullString

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, category, color, created_at
		FROM tags WHERE id = $1
	`, id).Scan(&tag.ID, &tag.Name, &category, &color, &tag.CreatedAt)
	if err != nil {
		return nil, err
	}

	tag.Category = category.String
	tag.Color = color.String
	return tag, nil
}

// GetByName 根据名称获取 Tag
func (r *TagRepoPG) GetByName(ctx context.Context, name string) (*model.Tag, error) {
	tag := &model.Tag{}
	var category, color sql.NullString

	err := r.pool.QueryRow(ctx, `
		SELECT id, name, category, color, created_at
		FROM tags WHERE name = $1
	`, name).Scan(&tag.ID, &tag.Name, &category, &color, &tag.CreatedAt)
	if err != nil {
		return nil, err
	}

	tag.Category = category.String
	tag.Color = color.String
	return tag, nil
}

// List 列出所有 Tags
func (r *TagRepoPG) List(ctx context.Context, category string) ([]model.Tag, error) {
	query := "SELECT id, name, category, color, created_at FROM tags"
	args := []interface{}{}

	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}

	query += " ORDER BY name"

	rows, err := r.pool.Query(ctx, query, args...)
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
func (r *TagRepoPG) Update(ctx context.Context, tag *model.Tag) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE tags SET name = $1, category = $2, color = $3
		WHERE id = $4
	`, tag.Name, tag.Category, tag.Color, tag.ID)
	return err
}

// Delete 删除 Tag
func (r *TagRepoPG) Delete(ctx context.Context, id int64) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM tags WHERE id = $1", id)
	return err
}

// GetOrCreateByName 根据名称获取或创建 Tag
func (r *TagRepoPG) GetOrCreateByName(ctx context.Context, name string, category string) (*model.Tag, error) {
	tag, err := r.GetByName(ctx, name)
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
	if err := r.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

// GetItemCounts 获取每个标签关联的 Item 数量
func (r *TagRepoPG) GetItemCounts(ctx context.Context) (map[int64]int, error) {
	rows, err := r.pool.Query(ctx, `
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
