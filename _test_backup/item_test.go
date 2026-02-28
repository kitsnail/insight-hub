package repository

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/kitsnail/insight-hub/internal/model"
)

func setupTestDB(t *testing.T) *Database {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	return db
}

func TestItemRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	item := &model.Item{
		ID:      "test-item-1",
		Type:    model.ItemTypeNote,
		Title:   "Test Note",
		Content: "This is a test note",
		Source:  "manual",
	}

	err := repo.Create(item)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 验证 CreatedAt 和 UpdatedAt 已设置
	if item.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
	if item.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
	if item.Status != model.ItemStatusActive {
		t.Errorf("Status should be active by default, got: %s", item.Status)
	}
}

func TestItemRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建测试数据
	item := &model.Item{
		ID:      "test-item-2",
		Type:    model.ItemTypeNote,
		Title:   "Test Note for Get",
		Content: "Content for get test",
		Source:  "manual",
	}
	if err := repo.Create(item); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 测试获取
	got, err := repo.GetByID("test-item-2")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if got.Title != item.Title {
		t.Errorf("Title mismatch: got %s, want %s", got.Title, item.Title)
	}
	if got.Content != item.Content {
		t.Errorf("Content mismatch: got %s, want %s", got.Content, item.Content)
	}
}

func TestItemRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	_, err := repo.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent item")
	}
}

func TestItemRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建多个测试数据
	for i := 1; i <= 5; i++ {
		item := &model.Item{
			ID:      "test-item-list-" + string(rune('0'+i)),
			Type:    model.ItemTypeNote,
			Title:   "Test Note",
			Content: "Content",
			Source:  "manual",
		}
		// 稍微延迟以区分创建时间
		time.Sleep(time.Millisecond)
		if err := repo.Create(item); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// 测试分页
	items, total, err := repo.List(3, 0, "", "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 5 {
		t.Errorf("Total count mismatch: got %d, want 5", total)
	}
	if len(items) != 3 {
		t.Errorf("Items count mismatch: got %d, want 3", len(items))
	}
}

func TestItemRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建测试数据
	item := &model.Item{
		ID:      "test-item-update",
		Type:    model.ItemTypeNote,
		Title:   "Original Title",
		Content: "Original Content",
		Source:  "manual",
	}
	if err := repo.Create(item); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 更新
	item.Title = "Updated Title"
	item.Content = "Updated Content"
	if err := repo.Update(item); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 验证
	got, err := repo.GetByID("test-item-update")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Title != "Updated Title" {
		t.Errorf("Title not updated: got %s", got.Title)
	}
}

func TestItemRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建测试数据
	item := &model.Item{
		ID:      "test-item-delete",
		Type:    model.ItemTypeNote,
		Title:   "To Be Deleted",
		Content: "Content",
		Source:  "manual",
	}
	if err := repo.Create(item); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 软删除
	if err := repo.Delete("test-item-delete"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证状态变为 deleted
	got, err := repo.GetByID("test-item-delete")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Status != model.ItemStatusDeleted {
		t.Errorf("Status should be deleted, got: %s", got.Status)
	}
}

func TestItemRepository_Search(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建测试数据
	items := []*model.Item{
		{ID: "search-1", Type: model.ItemTypeNote, Title: "Go Programming", Content: "Learn Go language", Source: "manual"},
		{ID: "search-2", Type: model.ItemTypeNote, Title: "Python Basics", Content: "Learn Python", Source: "manual"},
		{ID: "search-3", Type: model.ItemTypeNote, Title: "Kubernetes Guide", Content: "K8s orchestration with Go", Source: "manual"},
	}

	for _, item := range items {
		if err := repo.Create(item); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
		// 更新全文索引
		if err := repo.UpdateFTS(item); err != nil {
			t.Fatalf("UpdateFTS failed: %v", err)
		}
	}

	// 测试搜索
	results, err := repo.Search("Go", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// 应该找到包含 "Go" 的两个条目
	if len(results) < 2 {
		t.Errorf("Expected at least 2 results, got %d", len(results))
	}
}

func TestItemRepository_SetItemTags(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	tagRepo := NewTagRepository(db)

	// 创建 item
	item := &model.Item{
		ID:      "tag-test-1",
		Type:    model.ItemTypeNote,
		Title:   "Test",
		Content: "Content",
		Source:  "manual",
	}
	if err := itemRepo.Create(item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 创建标签
	tag1 := &model.Tag{Name: "tag1"}
	tag2 := &model.Tag{Name: "tag2"}
	tag3 := &model.Tag{Name: "tag3"}
	tagRepo.Create(tag1)
	tagRepo.Create(tag2)
	tagRepo.Create(tag3)

	// 设置标签
	err := itemRepo.SetItemTags("tag-test-1", []int64{tag1.ID, tag2.ID})
	if err != nil {
		t.Fatalf("SetItemTags failed: %v", err)
	}

	// 验证
	got, _ := itemRepo.GetByID("tag-test-1")
	if len(got.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(got.Tags))
	}

	// 替换标签
	err = itemRepo.SetItemTags("tag-test-1", []int64{tag3.ID})
	if err != nil {
		t.Fatalf("SetItemTags (replace) failed: %v", err)
	}

	got, _ = itemRepo.GetByID("tag-test-1")
	if len(got.Tags) != 1 {
		t.Errorf("Expected 1 tag after replace, got %d", len(got.Tags))
	}
	if got.Tags[0].Name != "tag3" {
		t.Errorf("Tag should be tag3, got %s", got.Tags[0].Name)
	}
}

func TestItemRepository_RemoveItemTag(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	tagRepo := NewTagRepository(db)

	// 创建 item 带标签
	tag1 := &model.Tag{Name: "remove-test-1"}
	tag2 := &model.Tag{Name: "remove-test-2"}
	tagRepo.Create(tag1)
	tagRepo.Create(tag2)

	item := &model.Item{
		ID:      "remove-tag-test",
		Type:    model.ItemTypeNote,
		Title:   "Test",
		Content: "Content",
		Source:  "manual",
		Tags:    []model.Tag{*tag1, *tag2},
	}
	if err := itemRepo.Create(item); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 移除一个标签
	err := itemRepo.RemoveItemTag("remove-tag-test", tag1.ID)
	if err != nil {
		t.Fatalf("RemoveItemTag failed: %v", err)
	}

	// 验证只剩一个标签
	got, _ := itemRepo.GetByID("remove-tag-test")
	if len(got.Tags) != 1 {
		t.Errorf("Expected 1 tag after removal, got %d", len(got.Tags))
	}
	if got.Tags[0].Name != "remove-test-2" {
		t.Errorf("Wrong tag remaining: got %s", got.Tags[0].Name)
	}
}

func TestItemRepository_List_ByType(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建不同类型的 items
	repo.Create(&model.Item{ID: "type-1", Type: model.ItemTypeNote, Title: "Note", Source: "manual"})
	repo.Create(&model.Item{ID: "type-2", Type: model.ItemTypeNote, Title: "Note 2", Source: "manual"})
	repo.Create(&model.Item{ID: "type-3", Type: model.ItemTypeTask, Title: "Task", Source: "manual"})

	// 按类型过滤
	items, total, err := repo.List(10, 0, model.ItemTypeNote, "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected 2 notes, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}
}

func TestItemRepository_List_ByStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建 items
	repo.Create(&model.Item{ID: "status-1", Type: model.ItemTypeNote, Title: "Active", Source: "manual"})
	repo.Create(&model.Item{ID: "status-2", Type: model.ItemTypeNote, Title: "Deleted", Source: "manual", Status: model.ItemStatusDeleted})

	// 按状态过滤
	_, total, err := repo.List(10, 0, "", model.ItemStatusActive)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if total != 1 {
		t.Errorf("Expected 1 active item, got %d", total)
	}
}

func TestItemRepository_Search_ExcludeDeleted(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewItemRepository(db)

	// 创建 items
	item1 := &model.Item{ID: "search-del-1", Type: model.ItemTypeNote, Title: "Go Active", Content: "Go content", Source: "manual"}
	item2 := &model.Item{ID: "search-del-2", Type: model.ItemTypeNote, Title: "Go Deleted", Content: "Go deleted", Source: "manual", Status: model.ItemStatusDeleted}
	repo.Create(item1)
	repo.Create(item2)
	repo.UpdateFTS(item1)
	repo.UpdateFTS(item2)

	// 搜索应该排除已删除的
	results, err := repo.Search("Go", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, item := range results {
		if item.Status == model.ItemStatusDeleted {
			t.Error("Search should not return deleted items")
		}
	}
}
