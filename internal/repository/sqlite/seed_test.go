package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationSeedDataPopulatesCatalog(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	if err := sqliteRepo.SeedData(db); err != nil {
		t.Fatalf("SeedData: %v", err)
	}

	catRepo := sqliteRepo.NewCategoryRepo(db)
	cats, err := catRepo.List(ctx)
	if err != nil {
		t.Fatalf("List categories: %v", err)
	}
	if len(cats) != 5 {
		t.Fatalf("categories = %d, want 5", len(cats))
	}

	prodRepo := sqliteRepo.NewProductRepo(db)
	nProd, err := prodRepo.Count(ctx, 0)
	if err != nil {
		t.Fatalf("Count products: %v", err)
	}
	if nProd != 27 {
		t.Errorf("products = %d, want 27", nProd)
	}

	newsRepo := sqliteRepo.NewNewsRepo(db)
	nNews, err := newsRepo.Count(ctx)
	if err != nil {
		t.Fatalf("Count news: %v", err)
	}
	if nNews != 7 {
		t.Errorf("news = %d, want 7", nNews)
	}
}

func TestIntegrationSeedDataIdempotentWhenCategoriesExist(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	repo := sqliteRepo.NewCategoryRepo(db)
	if err := repo.Create(ctx, &entity.Category{Name: "Only", Slug: "only"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := sqliteRepo.SeedData(db); err != nil {
		t.Fatalf("SeedData: %v", err)
	}
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("after SeedData noop: %d categories, want 1", len(list))
	}
}
