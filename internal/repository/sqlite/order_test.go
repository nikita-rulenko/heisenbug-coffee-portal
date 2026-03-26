package sqlite_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func TestIntegrationOrderCRUD(t *testing.T) {
	db := setupTestDB(t)
	orderRepo := sqliteRepo.NewOrderRepo(db)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	p := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true}
	productRepo.Create(ctx, p)

	order := &entity.Order{
		CustomerID: "customer-1",
		Status:     entity.OrderStatusNew,
		Total:      700,
		Items: []entity.OrderItem{
			{ProductID: p.ID, Quantity: 2, Price: 350},
		},
	}

	if err := orderRepo.Create(ctx, order); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if order.ID == 0 {
		t.Fatal("expected non-zero order ID")
	}

	got, err := orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.CustomerID != "customer-1" {
		t.Errorf("CustomerID = %q, want %q", got.CustomerID, "customer-1")
	}
	if len(got.Items) != 1 {
		t.Fatalf("Items len = %d, want 1", len(got.Items))
	}
	if got.Items[0].Quantity != 2 {
		t.Errorf("Item quantity = %d, want 2", got.Items[0].Quantity)
	}
}

func TestIntegrationOrderStatusTransitions(t *testing.T) {
	db := setupTestDB(t)
	orderRepo := sqliteRepo.NewOrderRepo(db)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	p := &entity.Product{CategoryID: cat.ID, Name: "Раф", Price: 400, InStock: true}
	productRepo.Create(ctx, p)

	order := &entity.Order{
		CustomerID: "customer-2",
		Status:     entity.OrderStatusNew,
		Total:      400,
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1, Price: 400}},
	}
	orderRepo.Create(ctx, order)

	if err := orderRepo.UpdateStatus(ctx, order.ID, entity.OrderStatusProcessing); err != nil {
		t.Fatalf("UpdateStatus to processing: %v", err)
	}

	got, _ := orderRepo.GetByID(ctx, order.ID)
	if got.Status != entity.OrderStatusProcessing {
		t.Errorf("Status = %q, want %q", got.Status, entity.OrderStatusProcessing)
	}

	if err := orderRepo.UpdateStatus(ctx, order.ID, entity.OrderStatusCompleted); err != nil {
		t.Fatalf("UpdateStatus to completed: %v", err)
	}
}

func TestIntegrationOrderListByCustomer(t *testing.T) {
	db := setupTestDB(t)
	orderRepo := sqliteRepo.NewOrderRepo(db)
	productRepo := sqliteRepo.NewProductRepo(db)
	catRepo := sqliteRepo.NewCategoryRepo(db)
	ctx := context.Background()

	cat := seedCategory(t, catRepo)
	p := &entity.Product{CategoryID: cat.ID, Name: "Кофе", Price: 200, InStock: true}
	productRepo.Create(ctx, p)

	for i := range 3 {
		o := &entity.Order{
			CustomerID: "cust-A",
			Status:     entity.OrderStatusNew,
			Total:      float64(200 * (i + 1)),
			Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: i + 1, Price: 200}},
		}
		orderRepo.Create(ctx, o)
	}

	otherOrder := &entity.Order{
		CustomerID: "cust-B",
		Status:     entity.OrderStatusNew,
		Total:      200,
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1, Price: 200}},
	}
	orderRepo.Create(ctx, otherOrder)

	orders, err := orderRepo.ListByCustomer(ctx, "cust-A", 0, 100)
	if err != nil {
		t.Fatalf("ListByCustomer: %v", err)
	}
	if len(orders) != 3 {
		t.Errorf("cust-A orders = %d, want 3", len(orders))
	}

	orders, _ = orderRepo.ListByCustomer(ctx, "cust-B", 0, 100)
	if len(orders) != 1 {
		t.Errorf("cust-B orders = %d, want 1", len(orders))
	}
}

func TestIntegrationOrderGetNotFound(t *testing.T) {
	db := setupTestDB(t)
	orderRepo := sqliteRepo.NewOrderRepo(db)

	_, err := orderRepo.GetByID(context.Background(), 999)
	if err != entity.ErrNotFound {
		t.Errorf("GetByID non-existent = %v, want ErrNotFound", err)
	}
}
