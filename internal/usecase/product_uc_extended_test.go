package usecase_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitProductUCCreateWithAllFields(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p := &entity.Product{CategoryID: cat.ID, Name: "Латте", Description: "Молочный кофе", Price: 350, ImageURL: "/img/latte.jpg", InStock: true}
	if err := productUC.Create(ctx, p); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestUnitProductUCUpdateNonExistent(t *testing.T) {
	productUC, _, _, _ := setupUC(t)
	p := &entity.Product{ID: 999, Name: "Ghost", Price: 100, CategoryID: 1}
	err := productUC.Update(context.Background(), p)
	if err == nil {
		t.Error("expected error updating non-existent product")
	}
}

func TestUnitProductUCDeleteNonExistent(t *testing.T) {
	productUC, _, _, _ := setupUC(t)
	err := productUC.Delete(context.Background(), 999)
	if err == nil {
		t.Error("expected error deleting non-existent product")
	}
}

func TestUnitProductUCSearchEmptyQuery(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	seedCatAndProduct(t, productUC, categoryUC)

	results, err := productUC.Search(ctx, "", 10)
	if err != nil {
		t.Fatalf("Search(''): %v", err)
	}
	_ = results // just verify no error
}

func TestUnitProductUCSearchUnicode(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	productUC.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "Капучино", Price: 300, InStock: true})

	results, err := productUC.Search(ctx, "Капу", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("search 'Капу' = %d, want 1", len(results))
	}
}

func TestUnitProductUCListPaginationEdgeCases(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	for i := range 10 {
		productUC.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "P" + string(rune('A'+i)), Price: 100, InStock: true})
	}

	tests := []struct {
		name     string
		page     int
		pageSize int
		wantLen  int
	}{
		{"page 0 defaults", 0, 10, 10},
		{"pageSize 0 defaults to 20", 1, 0, 10},
		{"page past end", 100, 10, 0},
		{"pageSize 1", 1, 1, 1},
		{"pageSize 100", 1, 100, 10},
		{"page 2 size 3", 2, 3, 3},
		{"page 4 size 3", 4, 3, 1},
		{"page 1 size 10", 1, 10, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, _, err := productUC.List(ctx, 0, tt.page, tt.pageSize)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if len(items) != tt.wantLen {
				t.Errorf("List(0, %d, %d) = %d items, want %d", tt.page, tt.pageSize, len(items), tt.wantLen)
			}
		})
	}
}

func TestUnitProductUCListEmptyDB(t *testing.T) {
	productUC, _, _, _ := setupUC(t)
	items, total, err := productUC.List(context.Background(), 0, 1, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 0 || total != 0 {
		t.Errorf("List empty = %d items, total %d", len(items), total)
	}
}

func TestUnitProductUCCreateDuplicateName(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p1 := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true}
	productUC.Create(ctx, p1)
	p2 := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 400, InStock: true}
	err := productUC.Create(ctx, p2)
	if err != nil {
		t.Errorf("duplicate name should be allowed: %v", err)
	}
}

func TestUnitProductUCSearchNoResults(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	seedCatAndProduct(t, productUC, categoryUC)

	results, err := productUC.Search(ctx, "несуществующий_товар_xyz", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestUnitProductUCCreateAndRetrieve(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p := &entity.Product{CategoryID: cat.ID, Name: "Раф", Description: "Сливочный кофе", Price: 400, InStock: true}
	productUC.Create(ctx, p)

	got, err := productUC.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Раф" || got.Price != 400 || got.Description != "Сливочный кофе" {
		t.Errorf("fields mismatch: %+v", got)
	}
}
