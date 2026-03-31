package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func setupOrderTestData(t *testing.T) (*sqliteRepo.OrderRepo, *sqliteRepo.ProductRepo, *entity.Product) {
	t.Helper()
	db := setupTestDB(t)
	orderRepo := sqliteRepo.NewOrderRepo(db)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	p := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true}
	if err := productRepo.Create(ctx, p); err != nil {
		t.Fatalf("create product: %v", err)
	}
	return orderRepo, productRepo, p
}

func createTestOrder(t *testing.T, repo *sqliteRepo.OrderRepo, customerID string, product *entity.Product, qty int) *entity.Order {
	t.Helper()
	o := &entity.Order{
		CustomerID: customerID,
		Status:     entity.OrderStatusNew,
		Total:      product.Price * float64(qty),
		Items:      []entity.OrderItem{{ProductID: product.ID, Quantity: qty, Price: product.Price}},
	}
	if err := repo.Create(context.Background(), o); err != nil {
		t.Fatalf("create order: %v", err)
	}
	return o
}

func TestIntegrationOrderCreateMultipleItems(t *testing.T) {
	db := setupTestDB(t)
	orderRepo := sqliteRepo.NewOrderRepo(db)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	products := make([]*entity.Product, 5)
	for i := range 5 {
		p := &entity.Product{CategoryID: cat.ID, Name: "Product " + string(rune('A'+i)), Price: float64((i + 1) * 100), InStock: true}
		productRepo.Create(ctx, p)
		products[i] = p
	}

	items := make([]entity.OrderItem, 5)
	var total float64
	for i, p := range products {
		items[i] = entity.OrderItem{ProductID: p.ID, Quantity: i + 1, Price: p.Price}
		total += p.Price * float64(i+1)
	}

	order := &entity.Order{CustomerID: "cust-multi", Status: entity.OrderStatusNew, Total: total, Items: items}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.Items) != 5 {
		t.Errorf("Items count = %d, want 5", len(got.Items))
	}
}

func TestIntegrationOrderGetItemsSeparately(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	ctx := context.Background()

	order := &entity.Order{
		CustomerID: "cust-items",
		Status:     entity.OrderStatusNew,
		Total:      1050,
		Items: []entity.OrderItem{
			{ProductID: p.ID, Quantity: 2, Price: 350},
			{ProductID: p.ID, Quantity: 1, Price: 350},
		},
	}
	orderRepo.Create(ctx, order)

	items, err := orderRepo.GetItems(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetItems: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("items count = %d, want 2", len(items))
	}
}

func TestIntegrationOrderListByCustomerPagination(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	ctx := context.Background()

	for range 10 {
		createTestOrder(t, orderRepo, "cust-page", p, 1)
	}

	page1, _ := orderRepo.ListByCustomer(ctx, "cust-page", 0, 3)
	if len(page1) != 3 {
		t.Errorf("page1 = %d, want 3", len(page1))
	}
	page2, _ := orderRepo.ListByCustomer(ctx, "cust-page", 3, 3)
	if len(page2) != 3 {
		t.Errorf("page2 = %d, want 3", len(page2))
	}
	last, _ := orderRepo.ListByCustomer(ctx, "cust-page", 9, 5)
	if len(last) != 1 {
		t.Errorf("last page = %d, want 1", len(last))
	}
}

func TestIntegrationOrderListByCustomerEmpty(t *testing.T) {
	orderRepo, _, _ := setupOrderTestData(t)
	orders, err := orderRepo.ListByCustomer(context.Background(), "no-orders", 0, 100)
	if err != nil {
		t.Fatalf("ListByCustomer: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0, got %d", len(orders))
	}
}

func TestIntegrationOrderListByCustomerNonExistent(t *testing.T) {
	orderRepo, _, _ := setupOrderTestData(t)
	orders, err := orderRepo.ListByCustomer(context.Background(), "ghost-customer", 0, 10)
	if err != nil {
		t.Fatalf("ListByCustomer: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0, got %d", len(orders))
	}
}

func TestIntegrationOrderUpdateStatusNotFound(t *testing.T) {
	orderRepo, _, _ := setupOrderTestData(t)
	err := orderRepo.UpdateStatus(context.Background(), 999, entity.OrderStatusProcessing)
	if err != entity.ErrNotFound {
		t.Errorf("UpdateStatus non-existent = %v, want ErrNotFound", err)
	}
}

func TestIntegrationOrderUpdateStatusAllTransitions(t *testing.T) {
	tests := []struct {
		name   string
		to     entity.OrderStatus
	}{
		{"to processing", entity.OrderStatusProcessing},
		{"to completed", entity.OrderStatusCompleted},
		{"to cancelled", entity.OrderStatusCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo, _, p := setupOrderTestData(t)
			ctx := context.Background()
			o := createTestOrder(t, orderRepo, "cust-transition", p, 1)

			if tt.to == entity.OrderStatusCompleted {
				orderRepo.UpdateStatus(ctx, o.ID, entity.OrderStatusProcessing)
			}

			err := orderRepo.UpdateStatus(ctx, o.ID, tt.to)
			if err != nil {
				t.Fatalf("UpdateStatus to %s: %v", tt.to, err)
			}
			got, _ := orderRepo.GetByID(ctx, o.ID)
			if got.Status != tt.to {
				t.Errorf("status = %q, want %q", got.Status, tt.to)
			}
		})
	}
}

func TestIntegrationOrderCreateSetsTimestamps(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	o := createTestOrder(t, orderRepo, "cust-ts", p, 1)
	got, _ := orderRepo.GetByID(context.Background(), o.ID)
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestIntegrationOrderMultipleCustomers(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	ctx := context.Background()

	for range 2 {
		createTestOrder(t, orderRepo, "alice", p, 1)
	}
	for range 3 {
		createTestOrder(t, orderRepo, "bob", p, 1)
	}
	createTestOrder(t, orderRepo, "charlie", p, 1)

	alice, _ := orderRepo.ListByCustomer(ctx, "alice", 0, 100)
	bob, _ := orderRepo.ListByCustomer(ctx, "bob", 0, 100)
	charlie, _ := orderRepo.ListByCustomer(ctx, "charlie", 0, 100)

	if len(alice) != 2 {
		t.Errorf("alice = %d, want 2", len(alice))
	}
	if len(bob) != 3 {
		t.Errorf("bob = %d, want 3", len(bob))
	}
	if len(charlie) != 1 {
		t.Errorf("charlie = %d, want 1", len(charlie))
	}
}

func TestIntegrationOrderGetByIDReturnsItems(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	ctx := context.Background()

	order := &entity.Order{
		CustomerID: "cust-check-items",
		Status:     entity.OrderStatusNew,
		Total:      700,
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 2, Price: 350}},
	}
	orderRepo.Create(ctx, order)

	got, _ := orderRepo.GetByID(ctx, order.ID)
	if len(got.Items) == 0 {
		t.Fatal("GetByID should return items")
	}
	if got.Items[0].Price != 350 {
		t.Errorf("item price = %v, want 350", got.Items[0].Price)
	}
}

func TestIntegrationOrderLargeOrder(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	ctx := context.Background()

	items := make([]entity.OrderItem, 20)
	for i := range 20 {
		items[i] = entity.OrderItem{ProductID: p.ID, Quantity: i + 1, Price: p.Price}
	}

	order := &entity.Order{CustomerID: "cust-large", Status: entity.OrderStatusNew, Total: 73500, Items: items}
	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Create large order: %v", err)
	}

	got, _ := orderRepo.GetByID(ctx, order.ID)
	if len(got.Items) != 20 {
		t.Errorf("items = %d, want 20", len(got.Items))
	}
}

func TestIntegrationOrderCreateAndGetTotal(t *testing.T) {
	orderRepo, _, p := setupOrderTestData(t)
	ctx := context.Background()

	order := &entity.Order{
		CustomerID: "cust-total",
		Status:     entity.OrderStatusNew,
		Total:      1750,
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 5, Price: 350}},
	}
	orderRepo.Create(ctx, order)

	got, _ := orderRepo.GetByID(ctx, order.ID)
	if got.Total != 1750 {
		t.Errorf("total = %v, want 1750", got.Total)
	}
}
