package usecase_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitNewsUCCreateValidatesTitle(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	n := &entity.NewsItem{Title: "", Content: "Content"}
	if err := newsUC.Create(context.Background(), n); err != entity.ErrEmptyName {
		t.Errorf("Create empty title = %v, want ErrEmptyName", err)
	}
}

func TestUnitNewsUCCreateValidatesContent(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	n := &entity.NewsItem{Title: "Title", Content: ""}
	if err := newsUC.Create(context.Background(), n); err != entity.ErrEmptyContent {
		t.Errorf("Create empty content = %v, want ErrEmptyContent", err)
	}
}

func TestUnitNewsUCUpdateValidatesTitle(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()
	n := &entity.NewsItem{Title: "Title", Content: "Content"}
	newsUC.Create(ctx, n)

	n.Title = ""
	if err := newsUC.Update(ctx, n); err != entity.ErrEmptyName {
		t.Errorf("Update empty title = %v, want ErrEmptyName", err)
	}
}

func TestUnitNewsUCUpdateValidatesContent(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()
	n := &entity.NewsItem{Title: "Title", Content: "Content"}
	newsUC.Create(ctx, n)

	n.Content = ""
	if err := newsUC.Update(ctx, n); err != entity.ErrEmptyContent {
		t.Errorf("Update empty content = %v, want ErrEmptyContent", err)
	}
}

func TestUnitNewsUCUpdateNonExistent(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	n := &entity.NewsItem{ID: 999, Title: "Ghost", Content: "Content"}
	err := newsUC.Update(context.Background(), n)
	if err == nil {
		t.Error("expected error updating non-existent")
	}
}

func TestUnitNewsUCDeleteNonExistent(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	err := newsUC.Delete(context.Background(), 999)
	if err == nil {
		t.Error("expected error deleting non-existent")
	}
}

func TestUnitNewsUCGetByIDNonExistent(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	_, err := newsUC.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("GetByID 999 = %v, want ErrNotFound", err)
	}
}

func TestUnitNewsUCListEmpty(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	items, total, err := newsUC.List(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 || total != 0 {
		t.Errorf("List empty = %d items, total %d", len(items), total)
	}
}

func TestUnitNewsUCListPaginationEdgeCases(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	for i := range 10 {
		newsUC.Create(ctx, &entity.NewsItem{Title: "N" + string(rune('0'+i)), Content: "C"})
	}

	tests := []struct {
		name     string
		page     int
		pageSize int
		wantLen  int
		wantTotal int
	}{
		{"page 0 defaults", 0, 5, 5, 10},
		{"pageSize 0 defaults to 10", 1, 0, 10, 10},
		{"page past end", 100, 5, 0, 10},
		{"page 1 size 3", 1, 3, 3, 10},
		{"page 4 size 3", 4, 3, 1, 10},
		{"page 1 size 100", 1, 100, 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := newsUC.List(ctx, tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("List(%d, %d) = %d items, want %d", tt.page, tt.pageSize, len(items), tt.wantLen)
			}
			if total != tt.wantTotal {
				t.Errorf("total = %d, want %d", total, tt.wantTotal)
			}
		})
	}
}

func TestUnitNewsUCCreateWithAuthor(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()
	n := &entity.NewsItem{Title: "Тест", Content: "Контент", Author: "Автор"}
	if err := newsUC.Create(ctx, n); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, _ := newsUC.GetByID(ctx, n.ID)
	if got.Author != "Автор" {
		t.Errorf("author = %q, want Автор", got.Author)
	}
}

func TestUnitNewsUCCreateAndRetrieve(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()
	n := &entity.NewsItem{Title: "Новость", Content: "Текст новости", Author: "Редактор"}
	newsUC.Create(ctx, n)

	got, err := newsUC.GetByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Новость" || got.Content != "Текст новости" || got.Author != "Редактор" {
		t.Errorf("fields mismatch: %+v", got)
	}
}

func TestUnitNewsUCUpdateAllFields(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()
	n := &entity.NewsItem{Title: "Old", Content: "Old content", Author: "Old author"}
	newsUC.Create(ctx, n)

	n.Title = "New"
	n.Content = "New content"
	n.Author = "New author"
	newsUC.Update(ctx, n)

	got, _ := newsUC.GetByID(ctx, n.ID)
	if got.Title != "New" || got.Content != "New content" || got.Author != "New author" {
		t.Errorf("not updated: %+v", got)
	}
}

func TestUnitNewsUCListReturnsTotal(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	for i := range 15 {
		newsUC.Create(ctx, &entity.NewsItem{Title: "N" + string(rune('A'+i%26)), Content: "C"})
	}

	items, total, _ := newsUC.List(ctx, 1, 5)
	if len(items) != 5 {
		t.Errorf("items = %d, want 5", len(items))
	}
	if total != 15 {
		t.Errorf("total = %d, want 15", total)
	}
}
