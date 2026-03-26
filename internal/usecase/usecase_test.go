package usecase_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
	"github.com/nikita-rulenko/heisenbug-portal/internal/usecase"
)

func setupUC(t *testing.T) (*usecase.ProductUseCase, *usecase.CategoryUseCase, *usecase.OrderUseCase, *usecase.NewsUseCase) {
	t.Helper()
	db, err := sqliteRepo.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqliteRepo.RunMigrations(db)
	t.Cleanup(func() { db.Close() })

	productRepo := sqliteRepo.NewProductRepo(db)
	categoryRepo := sqliteRepo.NewCategoryRepo(db)
	orderRepo := sqliteRepo.NewOrderRepo(db)
	newsRepo := sqliteRepo.NewNewsRepo(db)

	return usecase.NewProductUseCase(productRepo, categoryRepo),
		usecase.NewCategoryUseCase(categoryRepo),
		usecase.NewOrderUseCase(orderRepo, productRepo),
		usecase.NewNewsUseCase(newsRepo)
}

func TestUnitProductUCCreateValidatesCategory(t *testing.T) {
	productUC, _, _, _ := setupUC(t)
	ctx := context.Background()

	p := &entity.Product{Name: "Латте", Price: 350, CategoryID: 999}
	err := productUC.Create(ctx, p)
	if err != entity.ErrInvalidCategory {
		t.Errorf("Create with non-existent category: err = %v, want ErrInvalidCategory", err)
	}
}

func TestUnitProductUCCreateValidatesProduct(t *testing.T) {
	productUC, _, _, _ := setupUC(t)
	ctx := context.Background()

	p := &entity.Product{Name: "", Price: 350, CategoryID: 1}
	err := productUC.Create(ctx, p)
	if err != entity.ErrEmptyName {
		t.Errorf("Create with empty name: err = %v, want ErrEmptyName", err)
	}
}

func TestUnitProductUCListPagination(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Тест", Slug: "test"}
	categoryUC.Create(ctx, cat)

	for i := range 25 {
		productUC.Create(ctx, &entity.Product{
			CategoryID: cat.ID,
			Name:       "Продукт " + string(rune('A'+i)),
			Price:      float64(100 + i*10),
		})
	}

	products, total, err := productUC.List(ctx, 0, 1, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(products) != 10 {
		t.Errorf("page 1 len = %d, want 10", len(products))
	}
	if total != 25 {
		t.Errorf("total = %d, want 25", total)
	}

	products, _, _ = productUC.List(ctx, 0, 3, 10)
	if len(products) != 5 {
		t.Errorf("page 3 len = %d, want 5", len(products))
	}
}

func TestUnitOrderUCCreateSetsPrice(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Напитки", Slug: "drinks"}
	categoryUC.Create(ctx, cat)

	p := &entity.Product{CategoryID: cat.ID, Name: "Раф", Price: 400, InStock: true}
	productUC.Create(ctx, p)

	order := &entity.Order{
		CustomerID: "test",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 3}},
	}

	if err := orderUC.Create(ctx, order); err != nil {
		t.Fatalf("Create order: %v", err)
	}

	if order.Items[0].Price != 400 {
		t.Errorf("item price = %v, want 400 (from product)", order.Items[0].Price)
	}
	if order.Total != 1200 {
		t.Errorf("total = %v, want 1200", order.Total)
	}
	if order.Status != entity.OrderStatusNew {
		t.Errorf("status = %q, want %q", order.Status, entity.OrderStatusNew)
	}
}

func TestUnitOrderUCCancelFlow(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p := &entity.Product{CategoryID: cat.ID, Name: "Американо", Price: 250}
	productUC.Create(ctx, p)

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, order)

	if err := orderUC.Cancel(ctx, order.ID); err != nil {
		t.Fatalf("Cancel new order: %v", err)
	}

	err := orderUC.Cancel(ctx, order.ID)
	if err != entity.ErrInvalidStatus {
		t.Errorf("Cancel cancelled order: err = %v, want ErrInvalidStatus", err)
	}
}

func TestUnitOrderUCCompleteRequiresProcessing(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350}
	productUC.Create(ctx, p)

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, order)

	err := orderUC.Complete(ctx, order.ID)
	if err != entity.ErrInvalidStatus {
		t.Errorf("Complete new order: err = %v, want ErrInvalidStatus", err)
	}

	orderUC.Process(ctx, order.ID)
	if err := orderUC.Complete(ctx, order.ID); err != nil {
		t.Fatalf("Complete after processing: %v", err)
	}
}

func TestUnitNewsUCPagination(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	for i := range 15 {
		newsUC.Create(ctx, &entity.NewsItem{
			Title:   "Новость " + string(rune('A'+i)),
			Content: "Контент новости",
			Author:  "Автор",
		})
	}

	items, total, err := newsUC.List(ctx, 1, 5)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 5 {
		t.Errorf("page 1 len = %d, want 5", len(items))
	}
	if total != 15 {
		t.Errorf("total = %d, want 15", total)
	}
}
