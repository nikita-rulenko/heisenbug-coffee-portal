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

func seedCatAndProduct(t *testing.T, productUC *usecase.ProductUseCase, categoryUC *usecase.CategoryUseCase) (*entity.Category, *entity.Product) {
	t.Helper()
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)
	p := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true}
	productUC.Create(ctx, p)
	return cat, p
}

// --- Product UseCase ---

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

func TestUnitProductUCCreateValidatesNegativePrice(t *testing.T) {
	productUC, _, _, _ := setupUC(t)
	ctx := context.Background()

	p := &entity.Product{Name: "Тест", Price: -1, CategoryID: 1}
	err := productUC.Create(ctx, p)
	if err != entity.ErrNegativePrice {
		t.Errorf("Create with negative price: err = %v, want ErrNegativePrice", err)
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

func TestUnitProductUCListDefaultPageSize(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Тест", Slug: "test"}
	categoryUC.Create(ctx, cat)

	for i := range 25 {
		productUC.Create(ctx, &entity.Product{
			CategoryID: cat.ID,
			Name:       "P" + string(rune('A'+i)),
			Price:      float64(100),
		})
	}

	products, _, _ := productUC.List(ctx, 0, 0, 0) // zero page/pageSize → defaults
	if len(products) != 20 {
		t.Errorf("default page size len = %d, want 20", len(products))
	}
}

func TestUnitProductUCGetByID(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	got, err := productUC.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Латте" {
		t.Errorf("Name = %q, want %q", got.Name, "Латте")
	}
}

func TestUnitProductUCGetByIDNotFound(t *testing.T) {
	productUC, _, _, _ := setupUC(t)

	_, err := productUC.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUnitProductUCUpdate(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	p.Name = "Раф"
	p.Price = 400
	if err := productUC.Update(ctx, p); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := productUC.GetByID(ctx, p.ID)
	if got.Name != "Раф" || got.Price != 400 {
		t.Errorf("Update not applied: %+v", got)
	}
}

func TestUnitProductUCUpdateValidation(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	p.Name = ""
	if err := productUC.Update(ctx, p); err != entity.ErrEmptyName {
		t.Errorf("Update empty name: err = %v, want ErrEmptyName", err)
	}
}

func TestUnitProductUCDelete(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	if err := productUC.Delete(ctx, p.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := productUC.GetByID(ctx, p.ID)
	if err != entity.ErrNotFound {
		t.Errorf("after delete: err = %v, want ErrNotFound", err)
	}
}

func TestUnitProductUCSearch(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	productUC.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "Капучино", Price: 300, Description: "с молоком"})
	productUC.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "Американо", Price: 250})

	results, err := productUC.Search(ctx, "Капучино", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Search len = %d, want 1", len(results))
	}
}

func TestUnitProductUCSearchDefaultLimit(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "T", Slug: "t"}
	categoryUC.Create(ctx, cat)

	for range 25 {
		productUC.Create(ctx, &entity.Product{CategoryID: cat.ID, Name: "Кофе", Price: 100})
	}

	results, _ := productUC.Search(ctx, "Кофе", 0) // 0 → default 20
	if len(results) != 20 {
		t.Errorf("Search default limit len = %d, want 20", len(results))
	}
}

func TestUnitProductUCListByCategory(t *testing.T) {
	productUC, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	cat1 := &entity.Category{Name: "Кофе", Slug: "coffee"}
	cat2 := &entity.Category{Name: "Чай", Slug: "tea"}
	categoryUC.Create(ctx, cat1)
	categoryUC.Create(ctx, cat2)

	productUC.Create(ctx, &entity.Product{CategoryID: cat1.ID, Name: "Латте", Price: 350})
	productUC.Create(ctx, &entity.Product{CategoryID: cat1.ID, Name: "Американо", Price: 250})
	productUC.Create(ctx, &entity.Product{CategoryID: cat2.ID, Name: "Зелёный чай", Price: 200})

	coffee, total, _ := productUC.List(ctx, cat1.ID, 1, 100)
	if len(coffee) != 2 || total != 2 {
		t.Errorf("coffee: len=%d total=%d, want 2,2", len(coffee), total)
	}

	tea, total, _ := productUC.List(ctx, cat2.ID, 1, 100)
	if len(tea) != 1 || total != 1 {
		t.Errorf("tea: len=%d total=%d, want 1,1", len(tea), total)
	}
}

// --- Order UseCase ---

func TestUnitOrderUCCreateSetsPrice(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	order := &entity.Order{
		CustomerID: "test",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 3}},
	}

	if err := orderUC.Create(ctx, order); err != nil {
		t.Fatalf("Create order: %v", err)
	}

	if order.Items[0].Price != 350 {
		t.Errorf("item price = %v, want 350 (from product)", order.Items[0].Price)
	}
	if order.Total != 1050 {
		t.Errorf("total = %v, want 1050", order.Total)
	}
	if order.Status != entity.OrderStatusNew {
		t.Errorf("status = %q, want %q", order.Status, entity.OrderStatusNew)
	}
}

func TestUnitOrderUCCreateMultipleItems(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p1 := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true}
	p2 := &entity.Product{CategoryID: cat.ID, Name: "Раф", Price: 400, InStock: true}
	productUC.Create(ctx, p1)
	productUC.Create(ctx, p2)

	order := &entity.Order{
		CustomerID: "cust",
		Items: []entity.OrderItem{
			{ProductID: p1.ID, Quantity: 2},
			{ProductID: p2.ID, Quantity: 1},
		},
	}
	orderUC.Create(ctx, order)

	if order.Total != 350*2+400 {
		t.Errorf("total = %v, want %v", order.Total, 350*2+400)
	}
}

func TestUnitOrderUCCreateInvalidProduct(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)
	ctx := context.Background()

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: 999, Quantity: 1}},
	}
	err := orderUC.Create(ctx, order)
	if err != entity.ErrInvalidProduct {
		t.Errorf("err = %v, want ErrInvalidProduct", err)
	}
}

func TestUnitOrderUCCreateEmptyCustomer(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)
	ctx := context.Background()

	order := &entity.Order{
		CustomerID: "",
		Items:      []entity.OrderItem{{ProductID: 1, Quantity: 1}},
	}
	err := orderUC.Create(ctx, order)
	if err != entity.ErrEmptyCustomerID {
		t.Errorf("err = %v, want ErrEmptyCustomerID", err)
	}
}

func TestUnitOrderUCCreateEmptyItems(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)
	ctx := context.Background()

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{},
	}
	err := orderUC.Create(ctx, order)
	if err != entity.ErrEmptyOrder {
		t.Errorf("err = %v, want ErrEmptyOrder", err)
	}
}

func TestUnitOrderUCGetByID(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, order)

	got, err := orderUC.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.CustomerID != "cust" {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID, "cust")
	}
}

func TestUnitOrderUCGetByIDNotFound(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)

	_, err := orderUC.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUnitOrderUCListByCustomer(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	for range 5 {
		orderUC.Create(ctx, &entity.Order{
			CustomerID: "alice",
			Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
		})
	}
	orderUC.Create(ctx, &entity.Order{
		CustomerID: "bob",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	})

	orders, err := orderUC.ListByCustomer(ctx, "alice", 1, 100)
	if err != nil {
		t.Fatalf("ListByCustomer: %v", err)
	}
	if len(orders) != 5 {
		t.Errorf("alice orders = %d, want 5", len(orders))
	}

	orders, _ = orderUC.ListByCustomer(ctx, "bob", 1, 100)
	if len(orders) != 1 {
		t.Errorf("bob orders = %d, want 1", len(orders))
	}
}

func TestUnitOrderUCListByCustomerPagination(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	for range 10 {
		orderUC.Create(ctx, &entity.Order{
			CustomerID: "cust",
			Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
		})
	}

	page1, _ := orderUC.ListByCustomer(ctx, "cust", 1, 3)
	page2, _ := orderUC.ListByCustomer(ctx, "cust", 2, 3)

	if len(page1) != 3 {
		t.Errorf("page1 len = %d, want 3", len(page1))
	}
	if len(page2) != 3 {
		t.Errorf("page2 len = %d, want 3", len(page2))
	}
}

func TestUnitOrderUCCancelFlow(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

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

func TestUnitOrderUCCancelProcessing(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, order)
	orderUC.Process(ctx, order.ID)

	if err := orderUC.Cancel(ctx, order.ID); err != nil {
		t.Fatalf("Cancel processing order: %v", err)
	}
}

func TestUnitOrderUCCancelCompleted(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, order)
	orderUC.Process(ctx, order.ID)
	orderUC.Complete(ctx, order.ID)

	err := orderUC.Cancel(ctx, order.ID)
	if err != entity.ErrInvalidStatus {
		t.Errorf("Cancel completed: err = %v, want ErrInvalidStatus", err)
	}
}

func TestUnitOrderUCCompleteRequiresProcessing(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

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

func TestUnitOrderUCProcessRequiresNew(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	order := &entity.Order{
		CustomerID: "cust",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, order)
	orderUC.Process(ctx, order.ID)

	err := orderUC.Process(ctx, order.ID)
	if err != entity.ErrInvalidStatus {
		t.Errorf("Process already processing: err = %v, want ErrInvalidStatus", err)
	}
}

func TestUnitOrderUCFullLifecycle(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()

	_, p := seedCatAndProduct(t, productUC, categoryUC)

	order := &entity.Order{
		CustomerID: "lifecycle-test",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 2}},
	}

	// Create
	if err := orderUC.Create(ctx, order); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, _ := orderUC.GetByID(ctx, order.ID)
	if got.Status != entity.OrderStatusNew {
		t.Errorf("status after create = %q", got.Status)
	}

	// Process
	orderUC.Process(ctx, order.ID)
	got, _ = orderUC.GetByID(ctx, order.ID)
	if got.Status != entity.OrderStatusProcessing {
		t.Errorf("status after process = %q", got.Status)
	}

	// Complete
	orderUC.Complete(ctx, order.ID)
	got, _ = orderUC.GetByID(ctx, order.ID)
	if got.Status != entity.OrderStatusCompleted {
		t.Errorf("status after complete = %q", got.Status)
	}
}

func TestUnitOrderUCCancelNotFound(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)

	err := orderUC.Cancel(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUnitOrderUCProcessNotFound(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)

	err := orderUC.Process(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUnitOrderUCCompleteNotFound(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)

	err := orderUC.Complete(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

// --- News UseCase ---

func TestUnitNewsUCCreate(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Акция", Content: "Скидка 50%", Author: "Админ"}
	if err := newsUC.Create(ctx, n); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if n.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestUnitNewsUCCreateValidationEmptyTitle(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	err := newsUC.Create(ctx, &entity.NewsItem{Title: "", Content: "X"})
	if err != entity.ErrEmptyName {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}

func TestUnitNewsUCCreateValidationEmptyContent(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	err := newsUC.Create(ctx, &entity.NewsItem{Title: "X", Content: ""})
	if err != entity.ErrEmptyContent {
		t.Errorf("err = %v, want ErrEmptyContent", err)
	}
}

func TestUnitNewsUCGetByID(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Тест", Content: "Контент"}
	newsUC.Create(ctx, n)

	got, err := newsUC.GetByID(ctx, n.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Title != "Тест" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestUnitNewsUCGetByIDNotFound(t *testing.T) {
	_, _, _, newsUC := setupUC(t)

	_, err := newsUC.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUnitNewsUCUpdate(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Старый", Content: "Контент"}
	newsUC.Create(ctx, n)

	n.Title = "Новый"
	if err := newsUC.Update(ctx, n); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := newsUC.GetByID(ctx, n.ID)
	if got.Title != "Новый" {
		t.Errorf("Title after update = %q", got.Title)
	}
}

func TestUnitNewsUCUpdateValidation(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "OK", Content: "OK"}
	newsUC.Create(ctx, n)

	n.Title = ""
	err := newsUC.Update(ctx, n)
	if err != entity.ErrEmptyName {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}

func TestUnitNewsUCDelete(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	n := &entity.NewsItem{Title: "Delete me", Content: "Content"}
	newsUC.Create(ctx, n)

	if err := newsUC.Delete(ctx, n.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := newsUC.GetByID(ctx, n.ID)
	if err != entity.ErrNotFound {
		t.Errorf("after delete: err = %v, want ErrNotFound", err)
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

func TestUnitNewsUCPaginationDefaults(t *testing.T) {
	_, _, _, newsUC := setupUC(t)
	ctx := context.Background()

	for range 15 {
		newsUC.Create(ctx, &entity.NewsItem{Title: "X", Content: "Y"})
	}

	items, total, _ := newsUC.List(ctx, 0, 0) // defaults
	if len(items) != 10 {
		t.Errorf("default page size len = %d, want 10", len(items))
	}
	if total != 15 {
		t.Errorf("total = %d, want 15", total)
	}
}

// --- Category UseCase ---

func TestUnitCategoryUCCreate(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	if err := categoryUC.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestUnitCategoryUCCreateValidationEmptyName(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	err := categoryUC.Create(ctx, &entity.Category{Name: "", Slug: "s"})
	if err != entity.ErrEmptyName {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}

func TestUnitCategoryUCCreateValidationEmptySlug(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	err := categoryUC.Create(ctx, &entity.Category{Name: "N", Slug: ""})
	if err != entity.ErrEmptySlug {
		t.Errorf("err = %v, want ErrEmptySlug", err)
	}
}

func TestUnitCategoryUCGetByID(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	c := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, c)

	got, err := categoryUC.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "Кофе" {
		t.Errorf("Name = %q", got.Name)
	}
}

func TestUnitCategoryUCGetBySlug(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	categoryUC.Create(ctx, &entity.Category{Name: "Чай", Slug: "tea"})

	got, err := categoryUC.GetBySlug(ctx, "tea")
	if err != nil {
		t.Fatalf("GetBySlug: %v", err)
	}
	if got.Name != "Чай" {
		t.Errorf("Name = %q", got.Name)
	}
}

func TestUnitCategoryUCGetBySlugNotFound(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)

	_, err := categoryUC.GetBySlug(context.Background(), "nope")
	if err != entity.ErrNotFound {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUnitCategoryUCList(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	categoryUC.Create(ctx, &entity.Category{Name: "A", Slug: "a"})
	categoryUC.Create(ctx, &entity.Category{Name: "B", Slug: "b"})

	list, err := categoryUC.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("len = %d, want 2", len(list))
	}
}

func TestUnitCategoryUCUpdate(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	c := &entity.Category{Name: "Old", Slug: "old"}
	categoryUC.Create(ctx, c)

	c.Name = "New"
	c.Slug = "new"
	if err := categoryUC.Update(ctx, c); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, _ := categoryUC.GetByID(ctx, c.ID)
	if got.Name != "New" {
		t.Errorf("Name after update = %q", got.Name)
	}
}

func TestUnitCategoryUCUpdateValidation(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	c := &entity.Category{Name: "OK", Slug: "ok"}
	categoryUC.Create(ctx, c)

	c.Name = ""
	err := categoryUC.Update(ctx, c)
	if err != entity.ErrEmptyName {
		t.Errorf("err = %v, want ErrEmptyName", err)
	}
}

func TestUnitCategoryUCDelete(t *testing.T) {
	_, categoryUC, _, _ := setupUC(t)
	ctx := context.Background()

	c := &entity.Category{Name: "Del", Slug: "del"}
	categoryUC.Create(ctx, c)

	if err := categoryUC.Delete(ctx, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := categoryUC.GetByID(ctx, c.ID)
	if err != entity.ErrNotFound {
		t.Errorf("after delete: err = %v, want ErrNotFound", err)
	}
}
