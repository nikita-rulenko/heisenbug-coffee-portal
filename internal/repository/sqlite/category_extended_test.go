package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationCategoryGetBySlugAfterUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	repo.Create(ctx, c)

	c.Slug = "coffee-new"
	repo.Update(ctx, c)

	got, err := repo.GetBySlug(ctx, "coffee-new")
	if err != nil {
		t.Fatalf("GetBySlug new: %v", err)
	}
	if got.Slug != "coffee-new" {
		t.Errorf("slug = %q, want coffee-new", got.Slug)
	}

	_, err = repo.GetBySlug(ctx, "coffee")
	if err != entity.ErrNotFound {
		t.Errorf("GetBySlug old = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryListSortOrderVerification(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	orders := []int{5, 1, 3, 2, 4}
	for i, o := range orders {
		c := &entity.Category{Name: "Cat" + string(rune('A'+i)), Slug: "cat-" + string(rune('a'+i)), SortOrder: o}
		repo.Create(ctx, c)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 5 {
		t.Fatalf("count = %d, want 5", len(list))
	}
	// Should be sorted by sort_order ASC
	for i := 1; i < len(list); i++ {
		if list[i].SortOrder < list[i-1].SortOrder {
			t.Errorf("sort order violated: [%d]=%d < [%d]=%d", i, list[i].SortOrder, i-1, list[i-1].SortOrder)
		}
	}
}

func TestIntegrationCategoryUpdateDescription(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee", Description: "Старое описание"}
	repo.Create(ctx, c)

	c.Description = "Новое описание"
	repo.Update(ctx, c)

	got, _ := repo.GetByID(ctx, c.ID)
	if got.Description != "Новое описание" {
		t.Errorf("description = %q, want 'Новое описание'", got.Description)
	}
}

func TestIntegrationCategoryCreateSetsTimestamp(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	repo.Create(ctx, c)

	got, _ := repo.GetByID(ctx, c.ID)
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestIntegrationCategoryUpdateMultipleFields(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Old", Slug: "old", Description: "old desc", SortOrder: 1}
	repo.Create(ctx, c)

	c.Name = "New"
	c.Slug = "new"
	c.Description = "new desc"
	c.SortOrder = 99
	repo.Update(ctx, c)

	got, _ := repo.GetByID(ctx, c.ID)
	if got.Name != "New" || got.Slug != "new" || got.Description != "new desc" || got.SortOrder != 99 {
		t.Errorf("fields not updated: %+v", got)
	}
}

func TestIntegrationCategoryListAfterDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cats := make([]*entity.Category, 3)
	for i := range 3 {
		c := &entity.Category{Name: "C" + string(rune('0'+i)), Slug: "c" + string(rune('0'+i))}
		repo.Create(ctx, c)
		cats[i] = c
	}

	repo.Delete(ctx, cats[1].ID)

	list, _ := repo.List(ctx)
	if len(list) != 2 {
		t.Errorf("after delete: count = %d, want 2", len(list))
	}
}

func TestIntegrationCategoryGetByIDAfterDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	repo.Create(ctx, c)
	repo.Delete(ctx, c.ID)

	_, err := repo.GetByID(ctx, c.ID)
	if err != entity.ErrNotFound {
		t.Errorf("GetByID after delete = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryCreateManyVerifyCount(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	for i := range 15 {
		c := &entity.Category{Name: "Cat" + string(rune('A'+i%26)), Slug: "cat-" + string(rune('a'+i%26)) + string(rune('0'+i/26))}
		repo.Create(ctx, c)
	}

	list, _ := repo.List(ctx)
	if len(list) != 15 {
		t.Errorf("count = %d, want 15", len(list))
	}
}
