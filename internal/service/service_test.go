package service

import (
	"path/filepath"
	"testing"

	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/kitsnail/insight-hub/internal/repository"
)

func setupTestServices(t *testing.T) (*ItemService, *TagService) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := repository.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	itemRepo := repository.NewItemRepository(db)
	tagRepo := repository.NewTagRepository(db)

	return NewItemService(itemRepo, tagRepo), NewTagService(tagRepo)
}

func TestItemService_CreateItem(t *testing.T) {
	itemSvc, _ := setupTestServices(t)

	input := CreateItemInput{
		Type:    model.ItemTypeNote,
		Title:   "Test Note",
		Content: "This is a test note content",
		Tags:    []string{"test", "example"},
	}

	item, err := itemSvc.CreateItem(input)
	if err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	if item.ID == "" {
		t.Error("Item ID should be generated")
	}
	if item.Title != input.Title {
		t.Errorf("Title mismatch: got %s, want %s", item.Title, input.Title)
	}
	if len(item.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(item.Tags))
	}
}

func TestItemService_GetItem(t *testing.T) {
	itemSvc, _ := setupTestServices(t)

	// 创建测试数据
	created, err := itemSvc.CreateItem(CreateItemInput{
		Type:    model.ItemTypeNote,
		Title:   "Test for Get",
		Content: "Content",
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 获取
	got, err := itemSvc.GetItem(created.ID)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}

	if got.Title != created.Title {
		t.Errorf("Title mismatch: got %s, want %s", got.Title, created.Title)
	}
}

func TestItemService_ListItems(t *testing.T) {
	itemSvc, _ := setupTestServices(t)

	// 创建多个条目
	for i := 1; i <= 25; i++ {
		_, err := itemSvc.CreateItem(CreateItemInput{
			Type:    model.ItemTypeNote,
			Title:   "Test Note",
			Content: "Content",
		})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// 测试分页
	output, err := itemSvc.ListItems(ListItemsInput{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}

	if output.Total != 25 {
		t.Errorf("Total count mismatch: got %d, want 25", output.Total)
	}
	if len(output.Items) != 10 {
		t.Errorf("Items count mismatch: got %d, want 10", len(output.Items))
	}

	// 测试第二页
	output2, err := itemSvc.ListItems(ListItemsInput{
		Page:     2,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("ListItems page 2 failed: %v", err)
	}
	if len(output2.Items) != 10 {
		t.Errorf("Page 2 items count mismatch: got %d, want 10", len(output2.Items))
	}
}

func TestItemService_UpdateItem(t *testing.T) {
	itemSvc, _ := setupTestServices(t)

	// 创建测试数据
	created, err := itemSvc.CreateItem(CreateItemInput{
		Type:    model.ItemTypeNote,
		Title:   "Original Title",
		Content: "Original Content",
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 更新
	newTitle := "Updated Title"
	newContent := "Updated Content"
	updated, err := itemSvc.UpdateItem(created.ID, UpdateItemInput{
		Title:   &newTitle,
		Content: &newContent,
	})
	if err != nil {
		t.Fatalf("UpdateItem failed: %v", err)
	}

	if updated.Title != newTitle {
		t.Errorf("Title not updated: got %s", updated.Title)
	}
	if updated.Content != newContent {
		t.Errorf("Content not updated: got %s", updated.Content)
	}
}

func TestItemService_DeleteItem(t *testing.T) {
	itemSvc, _ := setupTestServices(t)

	// 创建测试数据
	created, err := itemSvc.CreateItem(CreateItemInput{
		Type:    model.ItemTypeNote,
		Title:   "To Be Deleted",
		Content: "Content",
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 删除
	if err := itemSvc.DeleteItem(created.ID); err != nil {
		t.Fatalf("DeleteItem failed: %v", err)
	}

	// 验证状态
	got, err := itemSvc.GetItem(created.ID)
	if err != nil {
		t.Fatalf("GetItem failed: %v", err)
	}
	if got.Status != model.ItemStatusDeleted {
		t.Errorf("Status should be deleted, got: %s", got.Status)
	}
}

func TestItemService_SearchItems(t *testing.T) {
	itemSvc, _ := setupTestServices(t)

	// 创建测试数据
	items := []CreateItemInput{
		{Type: model.ItemTypeNote, Title: "Go Programming", Content: "Learn Go"},
		{Type: model.ItemTypeNote, Title: "Python Tutorial", Content: "Learn Python"},
		{Type: model.ItemTypeNote, Title: "Docker Guide", Content: "Containerization with Docker"},
	}

	for _, input := range items {
		if _, err := itemSvc.CreateItem(input); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// 搜索
	results, err := itemSvc.SearchItems("Go", 10)
	if err != nil {
		t.Fatalf("SearchItems failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least 1 search result")
	}
}

func TestTagService_CreateTag(t *testing.T) {
	_, tagSvc := setupTestServices(t)

	tag, err := tagSvc.CreateTag(CreateTagInput{
		Name:     "golang",
		Category: "language",
		Color:    "#00ADD8",
	})
	if err != nil {
		t.Fatalf("CreateTag failed: %v", err)
	}

	if tag.Name != "golang" {
		t.Errorf("Name mismatch: got %s", tag.Name)
	}
}

func TestTagService_GetTag(t *testing.T) {
	_, tagSvc := setupTestServices(t)

	// 创建测试数据
	created, err := tagSvc.CreateTag(CreateTagInput{
		Name:     "python",
		Category: "language",
		Color:    "#3776AB",
	})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 获取
	got, err := tagSvc.GetTag(created.ID)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if got.Name != "python" {
		t.Errorf("Name mismatch: got %s", got.Name)
	}
}

func TestTagService_ListTags(t *testing.T) {
	_, tagSvc := setupTestServices(t)

	// 创建多个标签
	for _, name := range []string{"go", "python", "rust", "typescript"} {
		if _, err := tagSvc.CreateTag(CreateTagInput{Name: name, Category: "language"}); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// 列出
	tags, err := tagSvc.ListTags("")
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}

	if len(tags) != 4 {
		t.Errorf("Expected 4 tags, got %d", len(tags))
	}
}

func TestTagService_DeleteTag(t *testing.T) {
	_, tagSvc := setupTestServices(t)

	// 创建测试数据
	created, err := tagSvc.CreateTag(CreateTagInput{Name: "to-delete"})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 删除
	if err := tagSvc.DeleteTag(created.ID); err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}

	// 验证
	_, err = tagSvc.GetTag(created.ID)
	if err == nil {
		t.Error("Tag should be deleted")
	}
}
