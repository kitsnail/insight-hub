package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitsnail/insight-hub/internal/model"
)

// ItemRepoV3 Item 仓储 v3 实现
type ItemRepoV3 struct {
	pool *pgxpool.Pool
}

// NewItemRepoV3 创建 Item 仓储 v3
func NewItemRepoV3(pool *pgxpool.Pool) *ItemRepoV3 {
	return &ItemRepoV3{pool: pool}
}

// Create 创建 Item
func (r *ItemRepoV3) Create(ctx context.Context, req *model.ItemCreate) (*model.Item, error) {
	now := time.Now()
	item := &model.Item{
		ID:           req.ID,
		Type:         req.Type,
		Category:     req.Category,
		SourceSystem: req.SourceSystem,
		SourceAgent:  req.SourceAgent,
		SourceURL:    req.SourceURL,
		Title:        req.Title,
		Content:      req.Content,
		Summary:      req.Summary,
		Metadata:     req.Metadata,
		Status:       req.Status,
		Priority:     req.Priority,
		OccurredAt:   req.OccurredAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 设置默认值
	if item.Status == "" {
		item.Status = model.ItemStatusActive
	}

	// 处理 JSONB 字段
	var metadataJSON interface{}
	if item.Metadata != nil {
		metadataJSON = item.Metadata
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO items (
			id, type, category, source_system, source_agent, source_url,
			title, content, summary, metadata, status, priority,
			occurred_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, item.ID, item.Type, item.Category, item.SourceSystem, item.SourceAgent, item.SourceURL,
		item.Title, item.Content, item.Summary, metadataJSON, item.Status, item.Priority,
		item.OccurredAt, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create item: %w", err)
	}

	// 处理标签
	if len(req.Tags) > 0 {
		if err := r.SetItemTags(ctx, item.ID, req.Tags); err != nil {
			return nil, fmt.Errorf("set item tags: %w", err)
		}
	}

	return item, nil
}

// GetByID 根据 ID 获取 Item
func (r *ItemRepoV3) GetByID(ctx context.Context, id string) (*model.Item, error) {
	item := &model.Item{}
	var metadata []byte

	err := r.pool.QueryRow(ctx, `
		SELECT id, type, category, source_system, source_agent, source_url,
			   title, content, summary, metadata, status, priority,
			   occurred_at, created_at, updated_at
		FROM items WHERE id = $1
	`, id).Scan(
		&item.ID, &item.Type, &item.Category, &item.SourceSystem, &item.SourceAgent, &item.SourceURL,
		&item.Title, &item.Content, &item.Summary, &metadata, &item.Status, &item.Priority,
		&item.OccurredAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get item by id: %w", err)
	}

	// 处理 metadata
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
			item.Metadata = nil
		}
	}

	// 加载标签
	tags, err := r.GetItemTags(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get item tags: %w", err)
	}
	item.Tags = tags

	return item, nil
}

// Update 更新 Item
func (r *ItemRepoV3) Update(ctx context.Context, id string, req *model.ItemUpdate) (*model.Item, error) {
	// 先获取现有 item
	item, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 应用更新
	if req.Type != nil {
		item.Type = *req.Type
	}
	if req.Category != nil {
		item.Category = req.Category
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Content != nil {
		item.Content = req.Content
	}
	if req.Summary != nil {
		item.Summary = req.Summary
	}
	if req.Metadata != nil {
		item.Metadata = req.Metadata
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if req.Priority != nil {
		item.Priority = *req.Priority
	}
	if req.OccurredAt != nil {
		item.OccurredAt = req.OccurredAt
	}

	item.UpdatedAt = time.Now()

	// 处理 JSONB 字段
	var metadataJSON interface{}
	if item.Metadata != nil {
		metadataJSON = item.Metadata
	}

	_, err = r.pool.Exec(ctx, `
		UPDATE items SET
			type = $1, category = $2, title = $3, content = $4, summary = $5,
			metadata = $6, status = $7, priority = $8, occurred_at = $9, updated_at = $10
		WHERE id = $11
	`, item.Type, item.Category, item.Title, item.Content, item.Summary,
		metadataJSON, item.Status, item.Priority, item.OccurredAt, item.UpdatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}

	// 更新标签
	if req.Tags != nil {
		if err := r.SetItemTags(ctx, id, req.Tags); err != nil {
			return nil, fmt.Errorf("set item tags: %w", err)
		}
		item.Tags, _ = r.GetItemTags(ctx, id)
	}

	return item, nil
}

// Delete 软删除 Item
func (r *ItemRepoV3) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE items SET status = $1, updated_at = $2 WHERE id = $3
	`, model.ItemStatusDeleted, time.Now(), id)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return nil
}

// List 列出 Items
func (r *ItemRepoV3) List(ctx context.Context, query *model.ItemQuery) (*model.ItemListResponse, error) {
	// 设置默认值
	if query.Limit == 0 {
		query.Limit = 100
	}
	if query.Status == "" {
		query.Status = "active"
	}
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}

	// 构建查询
	sql := "FROM items WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	// 基本过滤
	if query.Type != "" {
		sql += fmt.Sprintf(" AND type = $%d", argIdx)
		args = append(args, query.Type)
		argIdx++
	}
	if query.Category != "" {
		sql += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, query.Category)
		argIdx++
	}
	if query.SourceSystem != "" {
		sql += fmt.Sprintf(" AND source_system = $%d", argIdx)
		args = append(args, query.SourceSystem)
		argIdx++
	}
	if query.SourceAgent != "" {
		sql += fmt.Sprintf(" AND source_agent = $%d", argIdx)
		args = append(args, query.SourceAgent)
		argIdx++
	}
	if query.Status != "" {
		sql += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, query.Status)
		argIdx++
	}

	// 时间范围
	if query.DateFrom != "" {
		sql += fmt.Sprintf(" AND occurred_at >= $%d", argIdx)
		args = append(args, query.DateFrom)
		argIdx++
	}
	if query.DateTo != "" {
		sql += fmt.Sprintf(" AND occurred_at <= $%d", argIdx)
		args = append(args, query.DateTo)
		argIdx++
	}

	// 全文搜索
	if query.Search != "" {
		sql += fmt.Sprintf(" AND search_vector @@ plainto_tsquery($%d)", argIdx)
		args = append(args, query.Search)
		argIdx++
	}

	// 标签过滤
	if query.Tags != "" {
		sql += fmt.Sprintf(` AND id IN (
			SELECT item_id FROM item_tags it 
			JOIN tags t ON it.tag_id = t.id 
			WHERE t.name = ANY(string_to_array($%d, ','))
		)`, argIdx)
		args = append(args, query.Tags)
		argIdx++
	}

	// 获取总数
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+sql, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count items: %w", err)
	}

	// 排序
	sortOrder := "ASC"
	if query.SortDesc {
		sortOrder = "DESC"
	}
	sql += fmt.Sprintf(" ORDER BY %s %s", query.SortBy, sortOrder)

	// 分页
	sql += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, query.Limit, query.Offset)

	// 查询列表
	rows, err := r.pool.Query(ctx, `
		SELECT id, type, category, source_system, source_agent, source_url,
			   title, content, summary, metadata, status, priority,
			   occurred_at, created_at, updated_at
	`+sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := []model.Item{}
	for rows.Next() {
		var item model.Item
		var metadata []byte

		err := rows.Scan(
			&item.ID, &item.Type, &item.Category, &item.SourceSystem, &item.SourceAgent, &item.SourceURL,
			&item.Title, &item.Content, &item.Summary, &metadata, &item.Status, &item.Priority,
			&item.OccurredAt, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		if len(metadata) > 0 {
			json.Unmarshal(metadata, &item.Metadata)
		}

		items = append(items, item)
	}

	return &model.ItemListResponse{
		Items:   items,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
		HasMore: query.Offset+query.Limit < total,
	}, nil
}

// Search 搜索 Items
func (r *ItemRepoV3) Search(ctx context.Context, query string, limit int) ([]*model.Item, error) {
	if limit == 0 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, type, category, source_system, source_agent, source_url,
			   title, content, summary, metadata, status, priority,
			   occurred_at, created_at, updated_at
		FROM items
		WHERE status != $1
		AND search_vector @@ plainto_tsquery($2)
		ORDER BY ts_rank(search_vector, plainto_tsquery($2)) DESC
		LIMIT $3
	`, model.ItemStatusDeleted, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search items: %w", err)
	}
	defer rows.Close()

	items := []*model.Item{}
	for rows.Next() {
		item := &model.Item{}
		var metadata []byte

		err := rows.Scan(
			&item.ID, &item.Type, &item.Category, &item.SourceSystem, &item.SourceAgent, &item.SourceURL,
			&item.Title, &item.Content, &item.Summary, &metadata, &item.Status, &item.Priority,
			&item.OccurredAt, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		if len(metadata) > 0 {
			json.Unmarshal(metadata, &item.Metadata)
		}

		items = append(items, item)
	}

	return items, nil
}

// BatchCreate 批量创建 Items
func (r *ItemRepoV3) BatchCreate(ctx context.Context, items []*model.ItemCreate) ([]*model.Item, error) {
	result := make([]*model.Item, 0, len(items))

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		for _, req := range items {
			now := time.Now()
			item := &model.Item{
				ID:           req.ID,
				Type:         req.Type,
				Category:     req.Category,
				SourceSystem: req.SourceSystem,
				SourceAgent:  req.SourceAgent,
				SourceURL:    req.SourceURL,
				Title:        req.Title,
				Content:      req.Content,
				Summary:      req.Summary,
				Metadata:     req.Metadata,
				Status:       req.Status,
				Priority:     req.Priority,
				OccurredAt:   req.OccurredAt,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if item.Status == "" {
				item.Status = model.ItemStatusActive
			}

			var metadataJSON interface{}
			if item.Metadata != nil {
				metadataJSON = item.Metadata
			}

			_, err := tx.Exec(ctx, `
				INSERT INTO items (
					id, type, category, source_system, source_agent, source_url,
					title, content, summary, metadata, status, priority,
					occurred_at, created_at, updated_at
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
			`, item.ID, item.Type, item.Category, item.SourceSystem, item.SourceAgent, item.SourceURL,
				item.Title, item.Content, item.Summary, metadataJSON, item.Status, item.Priority,
				item.OccurredAt, item.CreatedAt, item.UpdatedAt)
			if err != nil {
				return fmt.Errorf("create item %s: %w", item.ID, err)
			}

			result = append(result, item)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// Stats 统计
func (r *ItemRepoV3) Stats(ctx context.Context) (*model.StatsResponse, error) {
	stats := &model.StatsResponse{
		ByType:   make(map[string]int),
		BySource: make(map[string]int),
		ByStatus: make(map[string]int),
		ByTag:    make(map[string]int),
	}

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM items WHERE status != $1
	`, model.ItemStatusDeleted).Scan(&stats.TotalItems)
	if err != nil {
		return nil, fmt.Errorf("count total: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT type, COUNT(*) FROM items 
		WHERE status != $1 
		GROUP BY type
	`, model.ItemStatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("count by type: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var typ string
		var count int
		if err := rows.Scan(&typ, &count); err != nil {
			return nil, err
		}
		stats.ByType[typ] = count
	}

	rows, err = r.pool.Query(ctx, `
		SELECT source_system, COUNT(*) FROM items 
		WHERE status != $1 
		GROUP BY source_system
	`, model.ItemStatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("count by source: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, err
		}
		stats.BySource[source] = count
	}

	rows, err = r.pool.Query(ctx, `
		SELECT status, COUNT(*) FROM items GROUP BY status
	`)
	if err != nil {
		return nil, fmt.Errorf("count by status: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.ByStatus[status] = count
	}

	rows, err = r.pool.Query(ctx, `
		SELECT t.name, COUNT(it.item_id) as count
		FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		JOIN items i ON it.item_id = i.id
		WHERE i.status != $1
		GROUP BY t.id, t.name
		ORDER BY count DESC
		LIMIT 50
	`, model.ItemStatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("count by tag: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, err
		}
		stats.ByTag[tag] = count
	}

	return stats, nil
}

// CountByType 按类型统计
func (r *ItemRepoV3) CountByType(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT type, COUNT(*) FROM items 
		WHERE status != $1 
		GROUP BY type
	`, model.ItemStatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("count by type: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var typ string
		var count int
		if err := rows.Scan(&typ, &count); err != nil {
			return nil, err
		}
		result[typ] = count
	}

	return result, nil
}

// CountBySource 按来源统计
func (r *ItemRepoV3) CountBySource(ctx context.Context) (map[string]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT source_system, COUNT(*) FROM items 
		WHERE status != $1 
		GROUP BY source_system
	`, model.ItemStatusDeleted)
	if err != nil {
		return nil, fmt.Errorf("count by source: %w", err)
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var source string
		var count int
		if err := rows.Scan(&source, &count); err != nil {
			return nil, err
		}
		result[source] = count
	}

	return result, nil
}

// SetItemTags 设置 Item 的标签
func (r *ItemRepoV3) SetItemTags(ctx context.Context, itemID string, tagNames []string) error {
	// 先清除现有标签
	_, err := r.pool.Exec(ctx, "DELETE FROM item_tags WHERE item_id = $1", itemID)
	if err != nil {
		return fmt.Errorf("clear item tags: %w", err)
	}

	// 为每个标签创建或获取 ID，然后关联
	for _, name := range tagNames {
		if name == "" {
			continue
		}

		// 获取或创建标签
		var tagID int64
		err := r.pool.QueryRow(ctx, `
			INSERT INTO tags (name) VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = $1
			RETURNING id
		`, name).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("create tag %s: %w", name, err)
		}

		// 关联标签
		_, err = r.pool.Exec(ctx, `
			INSERT INTO item_tags (item_id, tag_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, itemID, tagID)
		if err != nil {
			return fmt.Errorf("link tag %s: %w", name, err)
		}
	}

	return nil
}

// GetItemTags 获取 Item 的标签
func (r *ItemRepoV3) GetItemTags(ctx context.Context, itemID string) ([]model.Tag, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT t.id, t.name, t.category, t.color, t.created_at
		FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = $1
		ORDER BY t.name
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("get item tags: %w", err)
	}
	defer rows.Close()

	tags := []model.Tag{}
	for rows.Next() {
		var tag model.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Category, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// CheckDuplicate 检查重复项
// 去重策略：
// 1. 如果提供了 external_id，检查 metadata->>'external_id'
// 2. 如果提供了 title + source_system，检查标题相似度
// 3. 如果提供了 content，检查内容哈希
func (r *ItemRepoV3) CheckDuplicate(ctx context.Context, req *model.DedupCheckRequest) (*model.DedupCheckResponse, error) {
	resp := &model.DedupCheckResponse{Exists: false}

	// 策略 1: 检查 external_id
	if req.ExternalID != "" && req.SourceSystem != "" {
		var itemID, title string
		var occurredAt *time.Time

		err := r.pool.QueryRow(ctx, `
			SELECT id, title, occurred_at 
			FROM items 
			WHERE source_system = $1 
			AND metadata->>'external_id' = $2
			AND status != $3
			LIMIT 1
		`, req.SourceSystem, req.ExternalID, model.ItemStatusDeleted).Scan(&itemID, &title, &occurredAt)

		if err == nil {
			resp.Exists = true
			resp.ItemID = itemID
			resp.Title = title
			resp.DedupReason = "external_id"
			if occurredAt != nil {
				resp.OccurredAt = occurredAt.Format(time.RFC3339)
			}
			return resp, nil
		}
	}

	// 策略 2: 检查标题相似度（同一来源系统）
	if req.Title != "" && req.SourceSystem != "" {
		var itemID, title string
		var occurredAt *time.Time

		// 使用 pg_trgm 相似度匹配（需要设置相似度阈值）
		err := r.pool.QueryRow(ctx, `
			SELECT id, title, occurred_at 
			FROM items 
			WHERE source_system = $1 
			AND status != $2
			AND similarity(title, $3) > 0.9
			ORDER BY similarity(title, $3) DESC
			LIMIT 1
		`, req.SourceSystem, model.ItemStatusDeleted, req.Title).Scan(&itemID, &title, &occurredAt)

		if err == nil {
			resp.Exists = true
			resp.ItemID = itemID
			resp.Title = title
			resp.DedupReason = "title_similarity"
			if occurredAt != nil {
				resp.OccurredAt = occurredAt.Format(time.RFC3339)
			}
			return resp, nil
		}
	}

	// 策略 3: 检查 source_system + source_agent 组合（如果提供）
	if req.SourceAgent != "" && req.SourceSystem != "" {
		var itemID, title string
		var occurredAt *time.Time

		query := `
			SELECT id, title, occurred_at 
			FROM items 
			WHERE source_system = $1 
			AND source_agent = $2 
			AND status != $3
		`
		args := []interface{}{req.SourceSystem, req.SourceAgent, model.ItemStatusDeleted}

		// 如果有 external_id，也加入条件
		if req.ExternalID != "" {
			query += " AND metadata->>'external_id' = $4"
			args = append(args, req.ExternalID)
		}

		query += " LIMIT 1"

		err := r.pool.QueryRow(ctx, query, args...).Scan(&itemID, &title, &occurredAt)

		if err == nil {
			resp.Exists = true
			resp.ItemID = itemID
			resp.Title = title
			resp.DedupReason = "source_match"
			if occurredAt != nil {
				resp.OccurredAt = occurredAt.Format(time.RFC3339)
			}
			return resp, nil
		}
	}

	return resp, nil
}

// BatchDelete 批量删除
func (r *ItemRepoV3) BatchDelete(ctx context.Context, ids []string) (*model.BatchDeleteResponse, error) {
	resp := &model.BatchDeleteResponse{
		Total:  len(ids),
		Failed: []model.BatchDeleteError{},
	}

	for _, id := range ids {
		err := r.Delete(ctx, id)
		if err != nil {
			resp.Failed = append(resp.Failed, model.BatchDeleteError{
				ID:    id,
				Error: err.Error(),
			})
		}
	}

	resp.Succeeded = resp.Total - len(resp.Failed)
	return resp, nil
}

// BatchUpdateStatus 批量更新状态
func (r *ItemRepoV3) BatchUpdateStatus(ctx context.Context, ids []string, status model.ItemStatus) (*model.BatchUpdateStatusResponse, error) {
	resp := &model.BatchUpdateStatusResponse{
		Total:  len(ids),
		Failed: []model.BatchUpdateStatusError{},
	}

	for _, id := range ids {
		update := &model.ItemUpdate{
			Status: &status,
		}
		_, err := r.Update(ctx, id, update)
		if err != nil {
			resp.Failed = append(resp.Failed, model.BatchUpdateStatusError{
				ID:    id,
				Error: err.Error(),
			})
		}
	}

	resp.Succeeded = resp.Total - len(resp.Failed)
	return resp, nil
}
