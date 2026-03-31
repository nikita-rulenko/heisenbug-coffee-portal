package sqlite_test

import (
	"context"
	"testing"
	"time"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationNewsUpdateAllFields(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Old Title", Content: "Old Content", Author: "Old Author"}
	repo.Create(ctx, n)

	n.Title = "New Title"
	n.Content = "New Content"
	n.Author = "New Author"
	repo.Update(ctx, n)

	got, _ := repo.GetByID(ctx, n.ID)
	if got.Title != "New Title" || got.Content != "New Content" || got.Author != "New Author" {
		t.Errorf("fields not updated: %+v", got)
	}
}

func TestIntegrationNewsListLargePage(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	for i := range 50 {
		n := &entity.NewsItem{Title: "News " + string(rune('A'+i%26)), Content: "Content " + string(rune('0'+i%10))}
		repo.Create(ctx, n)
	}

	// Paginate through all
	total := 0
	for offset := 0; offset < 50; offset += 10 {
		page, _ := repo.List(ctx, offset, 10)
		total += len(page)
	}
	if total != 50 {
		t.Errorf("total through pages = %d, want 50", total)
	}
}

func TestIntegrationNewsCreateSetsTimestamps(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Test", Content: "Content"}
	repo.Create(ctx, n)

	got, _ := repo.GetByID(ctx, n.ID)
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if got.PublishedAt.IsZero() {
		t.Error("PublishedAt is zero")
	}
}

func TestIntegrationNewsListOffsetBeyondEnd(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	for range 10 {
		repo.Create(ctx, &entity.NewsItem{Title: "T", Content: "C"})
	}

	items, _ := repo.List(ctx, 100, 10)
	if len(items) != 0 {
		t.Errorf("offset beyond end = %d, want 0", len(items))
	}
}

func TestIntegrationNewsCountMultipleCreateDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	ids := make([]int64, 5)
	for i := range 5 {
		n := &entity.NewsItem{Title: "T" + string(rune('0'+i)), Content: "C"}
		repo.Create(ctx, n)
		ids[i] = n.ID
	}

	repo.Delete(ctx, ids[1])
	repo.Delete(ctx, ids[3])

	count, _ := repo.Count(ctx)
	if count != 3 {
		t.Errorf("count after delete 2 of 5 = %d, want 3", count)
	}
}

func TestIntegrationNewsListEmptyDB(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)

	items, err := repo.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0, got %d", len(items))
	}
}

func TestIntegrationNewsCreateVerifyFields(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Тест", Content: "Контент новости", Author: "Автор"}
	repo.Create(ctx, n)

	got, _ := repo.GetByID(ctx, n.ID)
	if got.Title != "Тест" {
		t.Errorf("title = %q", got.Title)
	}
	if got.Content != "Контент новости" {
		t.Errorf("content = %q", got.Content)
	}
	if got.Author != "Автор" {
		t.Errorf("author = %q", got.Author)
	}
}

func TestIntegrationNewsUpdateOnlyTitle(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Original", Content: "Content", Author: "Author"}
	repo.Create(ctx, n)

	n.Title = "Updated"
	repo.Update(ctx, n)

	got, _ := repo.GetByID(ctx, n.ID)
	if got.Title != "Updated" {
		t.Errorf("title = %q, want Updated", got.Title)
	}
	if got.Content != "Content" {
		t.Errorf("content changed: %q", got.Content)
	}
	if got.Author != "Author" {
		t.Errorf("author changed: %q", got.Author)
	}
}

func TestIntegrationNewsDeleteAndRecount(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	ids := make([]int64, 3)
	for i := range 3 {
		n := &entity.NewsItem{Title: "T" + string(rune('0'+i)), Content: "C"}
		repo.Create(ctx, n)
		ids[i] = n.ID
	}

	repo.Delete(ctx, ids[1])

	count, _ := repo.Count(ctx)
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
	items, _ := repo.List(ctx, 0, 100)
	if len(items) != 2 {
		t.Errorf("list = %d, want 2", len(items))
	}
}

func TestIntegrationNewsListOrderConsistency(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewNewsRepo(db)
	ctx := context.Background()

	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range 5 {
		n := &entity.NewsItem{
			Title:       "News " + string(rune('A'+i)),
			Content:     "Content",
			PublishedAt: base.Add(time.Duration(i) * 24 * time.Hour),
		}
		repo.Create(ctx, n)
	}

	items, _ := repo.List(ctx, 0, 10)
	if len(items) < 2 {
		t.Fatalf("not enough items: %d", len(items))
	}
	// Should be DESC by published_at
	for i := 1; i < len(items); i++ {
		if items[i].PublishedAt.After(items[i-1].PublishedAt) {
			t.Errorf("order violated: [%d]=%v after [%d]=%v", i, items[i].PublishedAt, i-1, items[i-1].PublishedAt)
		}
	}
}
