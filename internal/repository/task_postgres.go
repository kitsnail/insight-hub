package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kitsnail/insight-hub/internal/model"
)

// TaskRepoPG Task 仓储 (PostgreSQL 版本)
type TaskRepoPG struct {
	pool *pgxpool.Pool
}

// NewTaskRepoPG 创建 Task 仓储
func NewTaskRepoPG(pool *pgxpool.Pool) *TaskRepoPG {
	return &TaskRepoPG{pool: pool}
}

// Create 创建 Task
func (r *TaskRepoPG) Create(ctx context.Context, task *model.Task) error {
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now
	if task.Status == "" {
		task.Status = model.TaskStatusTodo
	}
	if task.Priority == "" {
		task.Priority = model.TaskPriorityMedium
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO tasks (id, item_id, title, description, status, priority, due_date, started_at, completed_at, created_at, updated_at, recurrence)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, task.ID, task.ItemID, task.Title, task.Description, task.Status, task.Priority, task.DueDate, task.StartedAt, task.CompletedAt, task.CreatedAt, task.UpdatedAt, task.Recurrence)
	return err
}

// GetByID 根据 ID 获取 Task
func (r *TaskRepoPG) GetByID(ctx context.Context, id string) (*model.Task, error) {
	task := &model.Task{}
	var itemID, description, recurrence sql.NullString
	var dueDate, startedAt, completedAt sql.NullTime

	err := r.pool.QueryRow(ctx, `
		SELECT id, item_id, title, description, status, priority, due_date, started_at, completed_at, created_at, updated_at, recurrence
		FROM tasks WHERE id = $1
	`, id).Scan(&task.ID, &itemID, &task.Title, &description, &task.Status, &task.Priority,
		&dueDate, &startedAt, &completedAt, &task.CreatedAt, &task.UpdatedAt, &recurrence)
	if err != nil {
		return nil, err
	}

	task.ItemID = itemID.String
	task.Description = description.String
	task.Recurrence = recurrence.String
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}
	if startedAt.Valid {
		task.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

// List 列出 Tasks
func (r *TaskRepoPG) List(ctx context.Context, status model.TaskStatus, limit, offset int) ([]*model.Task, int, error) {
	query := "FROM tasks WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += " AND status = $1"
		args = append(args, status)
		argIdx++
	}

	// 获取总数
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+query, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 获取列表
	query += " ORDER BY created_at DESC LIMIT $2 OFFSET $3"
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, "SELECT id, item_id, title, description, status, priority, due_date, created_at, updated_at "+query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tasks := []*model.Task{}
	for rows.Next() {
		task := &model.Task{}
		var itemID, description sql.NullString
		var dueDate sql.NullTime

		err := rows.Scan(&task.ID, &itemID, &task.Title, &description, &task.Status, &task.Priority, &dueDate, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}

		task.ItemID = itemID.String
		task.Description = description.String
		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

// Update 更新 Task
func (r *TaskRepoPG) Update(ctx context.Context, task *model.Task) error {
	task.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, `
		UPDATE tasks SET item_id = $1, title = $2, description = $3, status = $4, priority = $5, due_date = $6, started_at = $7, completed_at = $8, updated_at = $9, recurrence = $10
		WHERE id = $11
	`, task.ItemID, task.Title, task.Description, task.Status, task.Priority, task.DueDate, task.StartedAt, task.CompletedAt, task.UpdatedAt, task.Recurrence, task.ID)
	return err
}

// Delete 删除 Task
func (r *TaskRepoPG) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	return err
}

// GetByItemID 根据 ItemID 获取 Tasks
func (r *TaskRepoPG) GetByItemID(ctx context.Context, itemID string) ([]*model.Task, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, item_id, title, description, status, priority, due_date, started_at, completed_at, created_at, updated_at, recurrence
		FROM tasks WHERE item_id = $1
		ORDER BY created_at DESC
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []*model.Task{}
	for rows.Next() {
		task := &model.Task{}
		var itemID, description, recurrence sql.NullString
		var dueDate, startedAt, completedAt sql.NullTime

		err := rows.Scan(&task.ID, &itemID, &task.Title, &description, &task.Status, &task.Priority,
			&dueDate, &startedAt, &completedAt, &task.CreatedAt, &task.UpdatedAt, &recurrence)
		if err != nil {
			return nil, err
		}

		task.ItemID = itemID.String
		task.Description = description.String
		task.Recurrence = recurrence.String
		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}
		if startedAt.Valid {
			task.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
