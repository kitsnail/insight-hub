package repository

import (
	"database/sql"
	"time"

	"github.com/kitsnail/insight-hub/internal/model"
)

// ItemRepository Item 仓储
type ItemRepository struct {
	db *Database
}

// NewItemRepository 创建 Item 仓储
func NewItemRepository(db *Database) *ItemRepository {
	return &ItemRepository{db: db}
}

// Create 创建 Item
func (r *ItemRepository) Create(item *model.Item) error {
	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	if item.Status == "" {
		item.Status = model.ItemStatusActive
	}

	result, err := r.db.Exec(`
		INSERT INTO items (id, type, title, content, summary, source, source_url, source_metadata, occurred_at, created_at, updated_at, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.ID, item.Type, item.Title, item.Content, item.Summary, item.Source, item.SourceURL, item.SourceMetadata, item.OccurredAt, item.CreatedAt, item.UpdatedAt, item.Status)
	if err != nil {
		return err
	}

	// 更新全文索引
	if err := r.UpdateFTS(item); err != nil {
		return err
	}

	// 处理标签关联
	if len(item.Tags) > 0 {
		for _, tag := range item.Tags {
			if err := r.addItemTag(item.ID, tag.ID); err != nil {
				return err
			}
		}
	}

	_ = result // ID 已在客户端生成
	return nil
}

// GetByID 根据 ID 获取 Item
func (r *ItemRepository) GetByID(id string) (*model.Item, error) {
	item := &model.Item{}
	var occurredAt sql.NullTime
	var sourceURL, sourceMetadata sql.NullString

	err := r.db.QueryRow(`
		SELECT id, type, title, content, summary, source, source_url, source_metadata, occurred_at, created_at, updated_at, status
		FROM items WHERE id = ?
	`, id).Scan(
		&item.ID, &item.Type, &item.Title, &item.Content, &item.Summary, &item.Source,
		&sourceURL, &sourceMetadata, &occurredAt,
		&item.CreatedAt, &item.UpdatedAt, &item.Status,
	)
	if err != nil {
		return nil, err
	}

	item.SourceURL = sourceURL.String
	item.SourceMetadata = sourceMetadata.String
	if occurredAt.Valid {
		item.OccurredAt = &occurredAt.Time
	}

	// 加载标签
	tags, err := r.getItemTags(id)
	if err != nil {
		return nil, err
	}
	item.Tags = tags

	return item, nil
}

// List 列出 Items
func (r *ItemRepository) List(limit, offset int, itemType model.ItemType, status model.ItemStatus) ([]*model.Item, int, error) {
	// 构建查询
	query := "FROM items WHERE 1=1"
	args := []interface{}{}

	if itemType != "" {
		query += " AND type = ?"
		args = append(args, itemType)
	}
	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}

	// 获取总数
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) "+query, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query("SELECT id, type, title, content, summary, source, source_url, occurred_at, created_at, updated_at, status "+query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		item := &model.Item{}
		var occurredAt sql.NullTime
		var sourceURL sql.NullString

		err := rows.Scan(
			&item.ID, &item.Type, &item.Title, &item.Content, &item.Summary, &item.Source,
			&sourceURL, &occurredAt,
			&item.CreatedAt, &item.UpdatedAt, &item.Status,
		)
		if err != nil {
			return nil, 0, err
		}

		item.SourceURL = sourceURL.String
		if occurredAt.Valid {
			item.OccurredAt = &occurredAt.Time
		}

		items = append(items, item)
	}

	return items, total, nil
}

// Update 更新 Item
func (r *ItemRepository) Update(item *model.Item) error {
	item.UpdatedAt = time.Now()

	_, err := r.db.Exec(`
		UPDATE items SET type = ?, title = ?, content = ?, summary = ?, source = ?, source_url = ?, source_metadata = ?, occurred_at = ?, updated_at = ?, status = ?
		WHERE id = ?
	`, item.Type, item.Title, item.Content, item.Summary, item.Source, item.SourceURL, item.SourceMetadata, item.OccurredAt, item.UpdatedAt, item.Status, item.ID)
	if err != nil {
		return err
	}

	// 更新全文索引
	return r.UpdateFTS(item)
}

// Delete 软删除 Item
func (r *ItemRepository) Delete(id string) error {
	_, err := r.db.Exec("UPDATE items SET status = ?, updated_at = ? WHERE id = ?", model.ItemStatusDeleted, time.Now(), id)
	if err != nil {
		return err
	}

	// 从全文索引删除
	_, err = r.db.Exec("DELETE FROM fts_items WHERE id = ?", id)
	return err
}

// Search 搜索 Items
func (r *ItemRepository) Search(query string, limit int) ([]*model.Item, error) {
	// FTS5 查询语法：用双引号包裹用户输入，避免特殊字符问题
	ftsQuery := `"` + query + `"`

	rows, err := r.db.Query(`
		SELECT i.id, i.type, i.title, i.content, i.summary, i.source, i.source_url, i.occurred_at, i.created_at, i.updated_at, i.status
		FROM items i
		JOIN fts_items f ON f.id = i.id
		WHERE fts_items MATCH ?
		AND i.status != ?
		ORDER BY bm25(fts_items)
		LIMIT ?
	`, ftsQuery, model.ItemStatusDeleted, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		item := &model.Item{}
		var occurredAt sql.NullTime
		var sourceURL sql.NullString

		err := rows.Scan(
			&item.ID, &item.Type, &item.Title, &item.Content, &item.Summary, &item.Source,
			&sourceURL, &occurredAt,
			&item.CreatedAt, &item.UpdatedAt, &item.Status,
		)
		if err != nil {
			return nil, err
		}

		item.SourceURL = sourceURL.String
		if occurredAt.Valid {
			item.OccurredAt = &occurredAt.Time
		}

		items = append(items, item)
	}

	return items, nil
}

// addItemTag 添加 Item-Tag 关联
func (r *ItemRepository) addItemTag(itemID string, tagID int64) error {
	_, err := r.db.Exec("INSERT OR IGNORE INTO item_tags (item_id, tag_id) VALUES (?, ?)", itemID, tagID)
	return err
}

// RemoveItemTag 移除 Item-Tag 关联
func (r *ItemRepository) RemoveItemTag(itemID string, tagID int64) error {
	_, err := r.db.Exec("DELETE FROM item_tags WHERE item_id = ? AND tag_id = ?", itemID, tagID)
	return err
}

// getItemTags 获取 Item 的标签
func (r *ItemRepository) getItemTags(itemID string) ([]model.Tag, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.name, t.category, t.color, t.created_at
		FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = ?
		ORDER BY t.name
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := []model.Tag{}
	for rows.Next() {
		var tag model.Tag
		var category, color sql.NullString
		err := rows.Scan(&tag.ID, &tag.Name, &category, &color, &tag.CreatedAt)
		if err != nil {
			return nil, err
		}
		tag.Category = category.String
		tag.Color = color.String
		tags = append(tags, tag)
	}

	return tags, nil
}

// SetItemTags 设置 Item 的标签（替换）
func (r *ItemRepository) SetItemTags(itemID string, tagIDs []int64) error {
	// 先清除现有标签
	_, err := r.db.Exec("DELETE FROM item_tags WHERE item_id = ?", itemID)
	if err != nil {
		return err
	}

	// 添加新标签
	for _, tagID := range tagIDs {
		if err := r.addItemTag(itemID, tagID); err != nil {
			return err
		}
	}

	return nil
}

// UpdateFTS 更新全文索引
func (r *ItemRepository) UpdateFTS(item *model.Item) error {
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO fts_items (id, title, content, summary, source)
		VALUES (?, ?, ?, ?, ?)
	`, item.ID, item.Title, item.Content, item.Summary, item.Source)
	return err
}
