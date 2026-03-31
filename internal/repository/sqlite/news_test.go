package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationNewsCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{
		Title:   "Открытие",
		Content: "Мы открылись!",
		Author:  "Админ",
	}
	if err := repo.Create(ctx, n); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if n.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if n.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	got, err := repo.GetByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Открытие" {
		t.Errorf("Title = %q, want %q", got.Title, "Открытие")
	}
	if got.Author != "Админ" {
		t.Errorf("Author = %q, want %q", got.Author, "Админ")
	}

	got.Title = "Обновление"
	got.Content = "Новый контент"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := repo.GetByID(ctx, n.ID)
	if updated.Title != "Обновление" {
		t.Errorf("updated Title = %q, want %q", updated.Title, "Обновление")
	}

	if err := repo.Delete(ctx, n.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, n.ID)
	if err != entity.ErrNotFound {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationNewsGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)

	_, err := repo.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationNewsListPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	for i := range 20 {
		repo.Create(ctx, &entity.NewsItem{
			Title:   "Новость " + string(rune('A'+i)),
			Content: "Контент",
			Author:  "Автор",
		})
	}

	page1, err := repo.List(ctx, 0, 5)
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if len(page1) != 5 {
		t.Errorf("page 1 len = %d, want 5", len(page1))
	}

	page2, err := repo.List(ctx, 5, 5)
	if err != nil {
		t.Fatalf("List page 2: %v", err)
	}
	if len(page2) != 5 {
		t.Errorf("page 2 len = %d, want 5", len(page2))
	}

	// pages should not overlap
	if page1[0].ID == page2[0].ID {
		t.Error("pages overlap")
	}
}

func TestIntegrationNewsListEmpty(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)

	items, err := repo.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("len = %d, want 0", len(items))
	}
}

func TestIntegrationNewsCount(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	count, _ := repo.Count(ctx)
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	for range 7 {
		repo.Create(ctx, &entity.NewsItem{Title: "X", Content: "Y"})
	}

	count, _ = repo.Count(ctx)
	if count != 7 {
		t.Errorf("count = %d, want 7", count)
	}
}

func TestIntegrationNewsUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)

	err := repo.Update(context.Background(), &entity.NewsItem{ID: 999, Title: "X", Content: "Y"})
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationNewsDeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)

	err := repo.Delete(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationNewsCreateWithPublishedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	pubTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	n := &entity.NewsItem{
		Title:       "Акция",
		Content:     "Летняя акция",
		PublishedAt: pubTime,
	}
	repo.Create(ctx, n)

	got, _ := repo.GetByID(ctx, n.ID)
	if got.PublishedAt.Year() != 2025 || got.PublishedAt.Month() != 6 {
		t.Errorf("PublishedAt = %v, want June 2025", got.PublishedAt)
	}
}

func TestIntegrationNewsCreateAutoPublishedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Авто", Content: "Тест"}
	before := time.Now().Add(-time.Second)
	repo.Create(ctx, n)

	got, _ := repo.GetByID(ctx, n.ID)
	if got.PublishedAt.Before(before) {
		t.Error("PublishedAt should be auto-set to now")
	}
}

func TestIntegrationNewsListOrderByPublishedDesc(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	for i := range 5 {
		n := &entity.NewsItem{
			Title:       "Новость",
			Content:     "Контент",
			PublishedAt: time.Now().Add(time.Duration(i) * time.Hour),
		}
		repo.Create(ctx, n)
	}

	items, _ := repo.List(ctx, 0, 10)
	if len(items) != 5 {
		t.Fatalf("len = %d, want 5", len(items))
	}
	// latest should be first
	for i := 1; i < len(items); i++ {
		if items[i].PublishedAt.After(items[i-1].PublishedAt) {
			t.Errorf("item %d published after item %d — not sorted DESC", i, i-1)
		}
	}
}

func TestIntegrationNewsCountAfterDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Delete me", Content: "Content"}
	repo.Create(ctx, n)

	count, _ := repo.Count(ctx)
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	repo.Delete(ctx, n.ID)
	count, _ = repo.Count(ctx)
	if count != 0 {
		t.Errorf("count after delete = %d, want 0", count)
	}
}
