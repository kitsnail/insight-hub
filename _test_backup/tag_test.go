package repository

import (
	"testing"

	"github.com/kitsnail/insight-hub/internal/model"
)

func TestTagRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	tag := &model.Tag{
		Name:     "golang",
		Category: "technology",
		Color:    "#00ADD8",
	}

	err := repo.Create(tag)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if tag.ID == 0 {
		t.Error("Tag ID should be set after creation")
	}
}

func TestTagRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 创建测试数据
	tag := &model.Tag{Name: "python"}
	if err := repo.Create(tag); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 获取
	got, err := repo.GetByID(tag.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if got.Name != "python" {
		t.Errorf("Name mismatch: got %s, want python", got.Name)
	}
}

func TestTagRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 创建测试数据
	tag := &model.Tag{Name: "kubernetes", Category: "tech"}
	if err := repo.Create(tag); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 按名称获取
	got, err := repo.GetByName("kubernetes")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}

	if got.ID != tag.ID {
		t.Errorf("ID mismatch: got %d, want %d", got.ID, tag.ID)
	}
}

func TestTagRepository_GetOrCreateByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 第一次创建
	tag1, err := repo.GetOrCreateByName("docker", "devops")
	if err != nil {
		t.Fatalf("GetOrCreateByName failed: %v", err)
	}
	if tag1.Name != "docker" {
		t.Errorf("Name mismatch: got %s", tag1.Name)
	}

	// 第二次应返回已存在的
	tag2, err := repo.GetOrCreateByName("docker", "different-category")
	if err != nil {
		t.Fatalf("GetOrCreateByName (second call) failed: %v", err)
	}
	if tag2.ID != tag1.ID {
		t.Error("Should return existing tag, not create new one")
	}
}

func TestTagRepository_List(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 创建多个标签
	tags := []model.Tag{
		{Name: "go", Category: "language"},
		{Name: "python", Category: "language"},
		{Name: "docker", Category: "tool"},
	}
	for _, tag := range tags {
		if err := repo.Create(&tag); err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// 列出所有
	got, err := repo.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(got))
	}

	// 按分类过滤
	langTags, err := repo.List("language")
	if err != nil {
		t.Fatalf("List with category failed: %v", err)
	}
	if len(langTags) != 2 {
		t.Errorf("Expected 2 language tags, got %d", len(langTags))
	}
}

func TestTagRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 创建测试数据
	tag := &model.Tag{Name: "to-delete"}
	if err := repo.Create(tag); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 删除
	if err := repo.Delete(tag.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 验证已删除
	_, err := repo.GetByID(tag.ID)
	if err == nil {
		t.Error("Tag should be deleted")
	}
}

func TestTagRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 创建测试数据
	tag := &model.Tag{Name: "original", Category: "test", Color: "#000000"}
	if err := repo.Create(tag); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// 更新
	tag.Name = "updated"
	tag.Category = "new-category"
	tag.Color = "#FFFFFF"
	if err := repo.Update(tag); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 验证
	got, _ := repo.GetByID(tag.ID)
	if got.Name != "updated" {
		t.Errorf("Name not updated: got %s", got.Name)
	}
	if got.Category != "new-category" {
		t.Errorf("Category not updated: got %s", got.Category)
	}
	if got.Color != "#FFFFFF" {
		t.Errorf("Color not updated: got %s", got.Color)
	}
}

func TestTagRepository_GetItemCounts(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	itemRepo := NewItemRepository(db)
	tagRepo := NewTagRepository(db)

	// 创建标签
	tag1 := &model.Tag{Name: "count-test-1"}
	tag2 := &model.Tag{Name: "count-test-2"}
	tag3 := &model.Tag{Name: "count-test-3"}
	tagRepo.Create(tag1)
	tagRepo.Create(tag2)
	tagRepo.Create(tag3)

	// 创建 items 并关联标签
	item1 := &model.Item{
		ID:      "count-item-1",
		Type:    model.ItemTypeNote,
		Title:   "Test 1",
		Source:  "manual",
		Tags:    []model.Tag{*tag1, *tag2},
	}
	item2 := &model.Item{
		ID:      "count-item-2",
		Type:    model.ItemTypeNote,
		Title:   "Test 2",
		Source:  "manual",
		Tags:    []model.Tag{*tag1},
	}
	itemRepo.Create(item1)
	itemRepo.Create(item2)
	// tag3 没有关联任何 item

	// 获取数量
	counts, err := tagRepo.GetItemCounts()
	if err != nil {
		t.Fatalf("GetItemCounts failed: %v", err)
	}

	if counts[tag1.ID] != 2 {
		t.Errorf("tag1 should have 2 items, got %d", counts[tag1.ID])
	}
	if counts[tag2.ID] != 1 {
		t.Errorf("tag2 should have 1 item, got %d", counts[tag2.ID])
	}
	if counts[tag3.ID] != 0 {
		t.Errorf("tag3 should have 0 items, got %d", counts[tag3.ID])
	}
}

func TestTagRepository_GetByName_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	_, err := repo.GetByName("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

func TestTagRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	_, err := repo.GetByID(99999)
	if err == nil {
		t.Error("Expected error for non-existent tag")
	}
}

func TestTagRepository_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewTagRepository(db)

	// 删除不存在的标签应该不报错（或报错，取决于实现）
	err := repo.Delete(99999)
	// SQLite DELETE 不存在的记录不会报错
	if err != nil {
		t.Logf("Delete non-existent returned error: %v", err)
	}
}
