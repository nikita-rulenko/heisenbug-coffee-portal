package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationCategoryCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee", Description: "Кофейные напитки", SortOrder: 1}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Кофе" {
		t.Errorf("Name = %q, want %q", got.Name, "Кофе")
	}
	if got.Slug != "coffee" {
		t.Errorf("Slug = %q, want %q", got.Slug, "coffee")
	}
	if got.SortOrder != 1 {
		t.Errorf("SortOrder = %d, want 1", got.SortOrder)
	}

	got.Name = "Чай"
	got.Slug = "tea"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := repo.GetByID(ctx, c.ID)
	if updated.Name != "Чай" {
		t.Errorf("updated Name = %q, want %q", updated.Name, "Чай")
	}

	if err := repo.Delete(ctx, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(ctx, c.ID)
	if err != entity.ErrNotFound {
		t.Errorf("after delete, err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryGetBySlug(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	repo.Create(ctx, &entity.Category{Name: "Десерты", Slug: "desserts"})

	got, err := repo.GetBySlug(ctx, "desserts")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.Name != "Десерты" {
		t.Errorf("Name = %q, want %q", got.Name, "Десерты")
	}
}

func TestIntegrationCategoryGetBySlugNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)

	_, err := repo.GetBySlug(context.Background(), "nonexistent")
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)

	_, err := repo.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryList(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	repo.Create(ctx, &entity.Category{Name: "Чай", Slug: "tea", SortOrder: 2})
	repo.Create(ctx, &entity.Category{Name: "Кофе", Slug: "coffee", SortOrder: 1})
	repo.Create(ctx, &entity.Category{Name: "Десерты", Slug: "desserts", SortOrder: 3})

	categories, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(categories) != 3 {
		t.Fatalf("len = %d, want 3", len(categories))
	}
	// should be sorted by sort_order
	if categories[0].Slug != "coffee" {
		t.Errorf("first category = %q, want coffee", categories[0].Slug)
	}
}

func TestIntegrationCategoryListEmpty(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)

	categories, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(categories) != 0 {
		t.Errorf("len = %d, want 0", len(categories))
	}
}

func TestIntegrationCategoryUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)

	err := repo.Update(context.Background(), &entity.Category{ID: 999, Name: "X", Slug: "x"})
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryDeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)

	err := repo.Delete(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationCategoryCreateMultiple(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	for i := range 10 {
		c := &entity.Category{
			Name:      "Категория " + string(rune('A'+i)),
			Slug:      "cat-" + string(rune('a'+i)),
			SortOrder: i,
		}
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	list, _ := repo.List(ctx)
	if len(list) != 10 {
		t.Errorf("len = %d, want 10", len(list))
	}
}

func TestIntegrationCategoryUpdateSortOrder(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee", SortOrder: 1}
	repo.Create(ctx, c)

	c.SortOrder = 99
	repo.Update(ctx, c)

	got, _ := repo.GetByID(ctx, c.ID)
	if got.SortOrder != 99 {
		t.Errorf("SortOrder = %d, want 99", got.SortOrder)
	}
}
