package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func seedCategory(t *testing.T, repo *sqliteRepo.CategoryRepo) *entity.Category {
	t.Helper()
	c := &entity.Category{Name: "Тестовая", Slug: "test-cat", Description: "Тестовая категория"}
	if err := repo.Create(context.Background(), c); err != nil {
		t.Fatalf("seed category: %v", err)
	}
	return c
}

func TestIntegrationProductCRUD(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)

	p := &entity.Product{
		CategoryID:  cat.ID,
		Name:        "Американо",
		Description: "Двойной эспрессо с водой",
		Price:       250,
		InStock:     true,
	}
	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := productRepo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Американо" {
		t.Errorf("Name = %q, want %q", got.Name, "Американо")
	}
	if got.Price != 250 {
		t.Errorf("Price = %v, want 250", got.Price)
	}

	got.Price = 280
	if err := productRepo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, _ := productRepo.GetByID(ctx, p.ID)
	if updated.Price != 280 {
		t.Errorf("updated Price = %v, want 280", updated.Price)
	}

	if err := productRepo.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = productRepo.GetByID(ctx, p.ID)
	if err != entity.ErrNotFound {
		t.Errorf("after delete, GetByID err = %v, want ErrNotFound", err)
	}
}

func TestIntegrationProductList(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)

	for i := range 5 {
		p := &entity.Product{
			CategoryID: cat.ID,
			Name:       "Продукт " + string(rune('A'+i)),
			Price:      float64(100 + i*50),
			InStock:    true,
		}
		productRepo.Create(ctx, p)
	}

	all, err := productRepo.List(ctx, 0, 0, 100)
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("List all len = %d, want 5", len(all))
	}

	byCat, err := productRepo.List(ctx, cat.ID, 0, 100)
	if err != nil {
		t.Fatalf("List by category: %v", err)
	}
	if len(byCat) != 5 {
		t.Errorf("List by cat len = %d, want 5", len(byCat))
	}

	page, err := productRepo.List(ctx, 0, 0, 2)
	if err != nil {
		t.Fatalf("List paginated: %v", err)
	}
	if len(page) != 2 {
		t.Errorf("paginated len = %d, want 2", len(page))
	}
}

func TestIntegrationProductSearch(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)

	products := []entity.Product{
		{CategoryID: cat.ID, Name: "Капучино", Description: "с молочной пенкой", Price: 320, InStock: true},
		{CategoryID: cat.ID, Name: "Латте", Description: "много молока", Price: 350, InStock: true},
		{CategoryID: cat.ID, Name: "Американо", Description: "без молока", Price: 250, InStock: true},
	}
	for i := range products {
		productRepo.Create(ctx, &products[i])
	}

	results, err := productRepo.Search(ctx, "молок", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Search 'молок' len = %d, want 2", len(results))
	}

	results, err = productRepo.Search(ctx, "Капучино", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search 'Капучино' len = %d, want 1", len(results))
	}
}

func TestIntegrationProductCount(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)

	count, _ := productRepo.Count(ctx, 0)
	if count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	for range 3 {
		productRepo.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "X", Price: 100, InStock: true})
	}

	count, _ = productRepo.Count(ctx, 0)
	if count != 3 {
		t.Errorf("count after insert = %d, want 3", count)
	}

	count, _ = productRepo.Count(ctx, cat.ID)
	if count != 3 {
		t.Errorf("count by category = %d, want 3", count)
	}

	count, _ = productRepo.Count(ctx, 999)
	if count != 0 {
		t.Errorf("count non-existent category = %d, want 0", count)
	}
}

func TestIntegrationProductDeleteNotFound(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)

	err := productRepo.Delete(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("Delete non-existent = %v, want ErrNotFound", err)
	}
}

func TestIntegrationProductGetByIDNotFound(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)

	_, err := productRepo.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("GetByID non-existent = %v, want ErrNotFound", err)
	}
}

func TestIntegrationProductUpdateNotFound(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)

	err := productRepo.Update(context.Background(), &entity.Product{ID: 999, Name: "X", Price: 100, CategoryID: 1})
	if err != entity.ErrNotFound {
		t.Errorf("Update non-existent = %v, want ErrNotFound", err)
	}
}

func TestIntegrationProductSearchNoResults(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	productRepo.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "Кофе", Price: 200})

	results, err := productRepo.Search(ctx, "несуществующий", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search no match len = %d, want 0", len(results))
	}
}

func TestIntegrationProductSearchEmptyDB(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)

	results, err := productRepo.Search(context.Background(), "test", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search empty DB len = %d, want 0", len(results))
	}
}

func TestIntegrationProductListWithOffset(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	for i := range 10 {
		productRepo.Create(ctx, &entity.Product{
			CategoryID: cat.ID,
			Name:       "P" + string(rune('0'+i)),
			Price:      float64(100 + i),
		})
	}

	page1, _ := productRepo.List(ctx, 0, 0, 3)
	page2, _ := productRepo.List(ctx, 0, 3, 3)
	page4, _ := productRepo.List(ctx, 0, 9, 3)

	if len(page1) != 3 {
		t.Errorf("page1 len = %d, want 3", len(page1))
	}
	if len(page2) != 3 {
		t.Errorf("page2 len = %d, want 3", len(page2))
	}
	if len(page4) != 1 {
		t.Errorf("last page len = %d, want 1", len(page4))
	}
	if page1[0].ID == page2[0].ID {
		t.Error("pages should not overlap")
	}
}

func TestIntegrationProductListByNonExistentCategory(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	productRepo.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "X", Price: 100})

	results, err := productRepo.List(ctx, 999, 0, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len = %d, want 0", len(results))
	}
}

func TestIntegrationProductSearchLimit(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	for range 10 {
		productRepo.Create(ctx, &entity.Product{
			CategoryID:  cat.ID,
			Name:        "Кофе вариант",
			Description: "кофе",
			Price:       200,
		})
	}

	results, _ := productRepo.Search(ctx, "кофе", 3)
	if len(results) != 3 {
		t.Errorf("Search with limit 3 len = %d, want 3", len(results))
	}
}

func TestIntegrationProductUpdateFields(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	p := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true, Description: "Молочный"}
	productRepo.Create(ctx, p)

	p.Name = "Раф"
	p.Price = 400
	p.InStock = false
	p.Description = "Ванильный"
	productRepo.Update(ctx, p)

	got, _ := productRepo.GetByID(ctx, p.ID)
	if got.Name != "Раф" || got.Price != 400 || got.InStock != false || got.Description != "Ванильный" {
		t.Errorf("Update fields mismatch: %+v", got)
	}
}

func TestIntegrationProductCountAfterDelete(t *testing.T) {
	db := setupTestDB(t)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	p := &entity.Product{CategoryID: cat.ID, Name: "X", Price: 100}
	productRepo.Create(ctx, p)

	count, _ := productRepo.Count(ctx, 0)
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	productRepo.Delete(ctx, p.ID)
	count, _ = productRepo.Count(ctx, 0)
	if count != 0 {
		t.Errorf("count after delete = %d, want 0", count)
	}
}
