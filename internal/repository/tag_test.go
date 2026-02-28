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
