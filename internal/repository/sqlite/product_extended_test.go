package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationProductCreateSetsTimestamps(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	cat := seedCategory(t, catRepo)

	p := &entity.Product{CategoryID: cat.ID, Name: "Тест", Price: 100, InStock: true}
	repo.Create(context.Background(), p)

	got, _ := repo.GetByID(context.Background(), p.ID)
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestIntegrationProductUpdatePreservesCreatedAt(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	cat := seedCategory(t, catRepo)
	ctx := context.Background()

	p := &entity.Product{CategoryID: cat.ID, Name: "Тест", Price: 100, InStock: true}
	repo.Create(ctx, p)
	original, _ := repo.GetByID(ctx, p.ID)

	original.Price = 200
	repo.Update(ctx, original)

	updated, _ := repo.GetByID(ctx, p.ID)
	if updated.CreatedAt.IsZero() {
		t.Error("CreatedAt became zero after update")
	}
}

func TestIntegrationProductListEmptyDB(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)

	products, err := repo.List(context.Background(), 0, 0, 100)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("expected 0, got %d", len(products))
	}
}

func TestIntegrationProductSearchByDescription(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	cat := seedCategory(t, catRepo)
	ctx := context.Background()

	p := &entity.Product{CategoryID: cat.ID, Name: "Капучино", Description: "Кофе с молочной пенкой", Price: 300, InStock: true}
	repo.Create(ctx, p)

	results, err := repo.Search(ctx, "пенк", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("search 'пенк' = %d, want 1", len(results))
	}
}

func TestIntegrationProductMultipleCategories(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cats := make([]*entity.Category, 3)
	for i := range 3 {
		c := &entity.Category{Name: "Cat" + string(rune('A'+i)), Slug: "cat-" + string(rune('a'+i))}
		catRepo.Create(ctx, c)
		cats[i] = c
	}

	for i, c := range cats {
		for j := range i + 2 {
			p := &entity.Product{CategoryID: c.ID, Name: c.Name + "-P" + string(rune('0'+j)), Price: 100, InStock: true}
			repo.Create(ctx, p)
		}
	}

	c0, _ := repo.Count(ctx, cats[0].ID)
	c1, _ := repo.Count(ctx, cats[1].ID)
	c2, _ := repo.Count(ctx, cats[2].ID)

	if c0 != 2 {
		t.Errorf("cat0 count = %d, want 2", c0)
	}
	if c1 != 3 {
		t.Errorf("cat1 count = %d, want 3", c1)
	}
	if c2 != 4 {
		t.Errorf("cat2 count = %d, want 4", c2)
	}
}

func TestIntegrationProductListPaginationBoundary(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	cat := seedCategory(t, catRepo)
	ctx := context.Background()

	for i := range 5 {
		p := &entity.Product{CategoryID: cat.ID, Name: "P" + string(rune('0'+i)), Price: 100, InStock: true}
		repo.Create(ctx, p)
	}

	tests := []struct {
		name   string
		offset int
		limit  int
		want   int
	}{
		{"offset past end", 10, 5, 0},
		{"offset=0 limit=1", 0, 1, 1},
		{"huge offset", 9999, 10, 0},
		{"offset=4 limit=100", 4, 100, 1},
		{"offset=0 limit=5", 0, 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := repo.List(ctx, 0, tt.offset, tt.limit)
			if len(got) != tt.want {
				t.Errorf("List(0, %d, %d) = %d, want %d", tt.offset, tt.limit, len(got), tt.want)
			}
		})
	}
}

func TestIntegrationProductSearchSpecialChars(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	cat := seedCategory(t, catRepo)
	ctx := context.Background()

	p := &entity.Product{CategoryID: cat.ID, Name: "Тест", Price: 100, InStock: true}
	repo.Create(ctx, p)

	tests := []struct {
		name  string
		query string
	}{
		{"percent", "%"},
		{"underscore", "_"},
		{"single quote", "'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic or error
			_, err := repo.Search(ctx, tt.query, 10)
			if err != nil {
				t.Errorf("Search(%q) error: %v", tt.query, err)
			}
		})
	}
}

func TestIntegrationProductCreateMultipleSameCategory(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	cat := seedCategory(t, catRepo)
	ctx := context.Background()

	for i := range 20 {
		p := &entity.Product{CategoryID: cat.ID, Name: "Product " + string(rune('A'+i%26)), Price: float64(i*10 + 100), InStock: true}
		repo.Create(ctx, p)
	}

	products, _ := repo.List(ctx, cat.ID, 0, 100)
	if len(products) != 20 {
		t.Errorf("list = %d, want 20", len(products))
	}
}

func TestIntegrationProductSearchEmptyQuery(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	ctx := context.Background()

	// Should not error on empty query
	_, err := repo.Search(ctx, "", 10)
	if err != nil {
		t.Errorf("Search('') error: %v", err)
	}
}

func TestIntegrationProductCountEmptyDB(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)

	count, err := repo.Count(context.Background(), 0)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestIntegrationProductCountMultipleCategories(t *testing.T) {
	db := setupTestDB(t)
	repo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat1 := &entity.Category{Name: "Кофе", Slug: "coffee"}
	catRepo.Create(ctx, cat1)
	cat2 := &entity.Category{Name: "Чай", Slug: "tea"}
	catRepo.Create(ctx, cat2)

	for range 3 {
		repo.Create(ctx, &entity.Product{CategoryID: cat1.ID, Name: "K", Price: 100, InStock: true})
	}
	for range 2 {
		repo.Create(ctx, &entity.Product{CategoryID: cat2.ID, Name: "T", Price: 50, InStock: true})
	}

	tests := []struct {
		name string
		catID int64
		want int
	}{
		{"all", 0, 5},
		{"coffee", cat1.ID, 3},
		{"tea", cat2.ID, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := repo.Count(ctx, tt.catID)
			if got != tt.want {
				t.Errorf("Count(%d) = %d, want %d", tt.catID, got, tt.want)
			}
		})
	}
}
