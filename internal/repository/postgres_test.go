package repository

import (
	"context"
	"testing"
	"time"

	"github.com/kitsnail/insight-hub/internal/config"
	"github.com/kitsnail/insight-hub/internal/model"
)

func TestPostgresDB_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5433,
		Name:            "insight_hub",
		User:            "insight",
		Password:        "insight_local_dev",
		SSLMode:         "disable",
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: "1h",
		MaxConnIdleTime: "10m",
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

func TestItemRepoPG_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5433,
		Name:            "insight_hub",
		User:            "insight",
		Password:        "insight_local_dev",
		SSLMode:         "disable",
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: "1h",
		MaxConnIdleTime: "10m",
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	itemRepo := NewItemRepoPG(db.Pool)

	// Create
	item := &model.Item{
		ID:      "test-" + time.Now().Format("20060102150405"),
		Type:    model.ItemTypeInsight,
		Agent:   model.AgentUnknown,
		Title:   "Test Item",
		Content: "Test content",
		Source:  "test",
		Status:  model.ItemStatusActive,
	}

	if err := itemRepo.Create(ctx, item); err != nil {
		t.Fatalf("Failed to create item: %v", err)
	}

	// Read
	got, err := itemRepo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get item: %v", err)
	}

	if got.Title != item.Title {
		t.Errorf("Title mismatch: got %s, want %s", got.Title, item.Title)
	}

	// Update
	got.Title = "Updated Title"
	if err := itemRepo.Update(ctx, got); err != nil {
		t.Fatalf("Failed to update item: %v", err)
	}

	updated, err := itemRepo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get updated item: %v", err)
	}

	if updated.Title != "Updated Title" {
		t.Errorf("Title not updated: got %s", updated.Title)
	}

	// Delete (soft delete)
	if err := itemRepo.Delete(ctx, item.ID); err != nil {
		t.Fatalf("Failed to delete item: %v", err)
	}

	// Verify soft deleted - should still exist but with deleted status
	deleted, err := itemRepo.GetByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("Failed to get deleted item: %v", err)
	}
	if deleted.Status != model.ItemStatusDeleted {
		t.Errorf("Item status should be deleted, got %s", deleted.Status)
	}
}

func TestTagRepoPG_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5433,
		Name:            "insight_hub",
		User:            "insight",
		Password:        "insight_local_dev",
		SSLMode:         "disable",
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: "1h",
		MaxConnIdleTime: "10m",
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	tagRepo := NewTagRepoPG(db.Pool)

	// Create
	tag := &model.Tag{
		Name:     "test-tag-" + time.Now().Format("20060102150405"),
		Category: "test",
		Color:    "#FF0000",
	}

	if err := tagRepo.Create(ctx, tag); err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	if tag.ID == 0 {
		t.Error("Tag ID should be set after creation")
	}

	// Read
	got, err := tagRepo.GetByID(ctx, tag.ID)
	if err != nil {
		t.Fatalf("Failed to get tag: %v", err)
	}

	if got.Name != tag.Name {
		t.Errorf("Name mismatch: got %s, want %s", got.Name, tag.Name)
	}

	// Update
	got.Color = "#00FF00"
	if err := tagRepo.Update(ctx, got); err != nil {
		t.Fatalf("Failed to update tag: %v", err)
	}

	// Delete
	if err := tagRepo.Delete(ctx, tag.ID); err != nil {
		t.Fatalf("Failed to delete tag: %v", err)
	}
}

func TestItemRepoPG_Search(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5433,
		Name:            "insight_hub",
		User:            "insight",
		Password:        "insight_local_dev",
		SSLMode:         "disable",
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: "1h",
		MaxConnIdleTime: "10m",
	}

	db, err := NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	itemRepo := NewItemRepoPG(db.Pool)

	// Create test items
	item1 := &model.Item{
		ID:      "search-test-1-" + time.Now().Format("20060102150405"),
		Type:    model.ItemTypeInsight,
		Agent:   model.AgentUnknown,
		Title:   "Golang Programming",
		Content: "Learn Go language",
		Source:  "test",
		Status:  model.ItemStatusActive,
	}
	item2 := &model.Item{
		ID:      "search-test-2-" + time.Now().Format("20060102150405"),
		Type:    model.ItemTypeInsight,
		Agent:   model.AgentUnknown,
		Title:   "Python Programming",
		Content: "Learn Python language",
		Source:  "test",
		Status:  model.ItemStatusActive,
	}

	if err := itemRepo.Create(ctx, item1); err != nil {
		t.Fatalf("Failed to create item1: %v", err)
	}
	if err := itemRepo.Create(ctx, item2); err != nil {
		t.Fatalf("Failed to create item2: %v", err)
	}

	// Search
	results, err := itemRepo.Search(ctx, "Golang", 10)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one search result")
	}

	// Cleanup
	itemRepo.Delete(ctx, item1.ID)
	itemRepo.Delete(ctx, item2.ID)
}
