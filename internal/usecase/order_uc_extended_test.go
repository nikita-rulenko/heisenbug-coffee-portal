package usecase_test

import (
	"context"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitOrderUCCreateMultipleItemsSameProduct(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC)

	o := &entity.Order{
		CustomerID: "cust-same",
		Items: []entity.OrderItem{
			{ProductID: p.ID, Quantity: 2},
			{ProductID: p.ID, Quantity: 3},
		},
	}
	if err := orderUC.Create(ctx, o); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// total = 350*2 + 350*3 = 1750
	if o.Total != 1750 {
		t.Errorf("total = %v, want 1750", o.Total)
	}
}

func TestUnitOrderUCCreateVerifiesEachProduct(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC)

	o := &entity.Order{
		CustomerID: "cust-bad",
		Items: []entity.OrderItem{
			{ProductID: p.ID, Quantity: 1},
			{ProductID: 99999, Quantity: 1},
		},
	}
	err := orderUC.Create(ctx, o)
	if err == nil {
		t.Error("expected error for non-existent product")
	}
}

func TestUnitOrderUCListByCustomerEmpty(t *testing.T) {
	_, _, orderUC, _ := setupUC(t)
	orders, err := orderUC.ListByCustomer(context.Background(), "nobody", 1, 20)
	if err != nil {
		t.Fatalf("ListByCustomer: %v", err)
	}
	if len(orders) != 0 {
		t.Errorf("expected 0, got %d", len(orders))
	}
}

func TestUnitOrderUCStatusTransitionMatrix(t *testing.T) {
	tests := []struct {
		name   string
		action string // "cancel", "process", "complete"
		setup  string // "new", "processing", "completed", "cancelled"
		wantOK bool
	}{
		{"cancel new", "cancel", "new", true},
		{"cancel processing", "cancel", "processing", true},
		{"cancel completed", "cancel", "completed", false},
		{"cancel cancelled", "cancel", "cancelled", false},
		{"process new", "process", "new", true},
		{"process processing", "process", "processing", false},
		{"process completed", "process", "completed", false},
		{"process cancelled", "process", "cancelled", false},
		{"complete new", "complete", "new", false},
		{"complete processing", "complete", "processing", true},
		{"complete completed", "complete", "completed", false},
		{"complete cancelled", "complete", "cancelled", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			productUC, categoryUC, orderUC, _ := setupUC(t)
			ctx := context.Background()
			_, p := seedCatAndProduct(t, productUC, categoryUC)

			o := &entity.Order{
				CustomerID: "cust-matrix",
				Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
			}
			orderUC.Create(ctx, o)

			// Setup state
			switch tt.setup {
			case "processing":
				orderUC.Process(ctx, o.ID)
			case "completed":
				orderUC.Process(ctx, o.ID)
				orderUC.Complete(ctx, o.ID)
			case "cancelled":
				orderUC.Cancel(ctx, o.ID)
			}

			// Try action
			var err error
			switch tt.action {
			case "cancel":
				err = orderUC.Cancel(ctx, o.ID)
			case "process":
				err = orderUC.Process(ctx, o.ID)
			case "complete":
				err = orderUC.Complete(ctx, o.ID)
			}

			if tt.wantOK && err != nil {
				t.Errorf("%s from %s: unexpected error %v", tt.action, tt.setup, err)
			}
			if !tt.wantOK && err == nil {
				t.Errorf("%s from %s: expected error", tt.action, tt.setup)
			}
		})
	}
}

func TestUnitOrderUCCreateSetsStatusNew(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC)

	o := &entity.Order{
		CustomerID: "cust-status",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, o)

	got, _ := orderUC.GetByID(ctx, o.ID)
	if got.Status != entity.OrderStatusNew {
		t.Errorf("status = %q, want new", got.Status)
	}
}

func TestUnitOrderUCCreateCalculatesCorrectTotal(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	cat := &entity.Category{Name: "Кофе", Slug: "coffee"}
	categoryUC.Create(ctx, cat)

	p1 := &entity.Product{CategoryID: cat.ID, Name: "Латте", Price: 350, InStock: true}
	productUC.Create(ctx, p1)
	p2 := &entity.Product{CategoryID: cat.ID, Name: "Раф", Price: 400, InStock: true}
	productUC.Create(ctx, p2)

	o := &entity.Order{
		CustomerID: "cust-calc",
		Items: []entity.OrderItem{
			{ProductID: p1.ID, Quantity: 2}, // 700
			{ProductID: p2.ID, Quantity: 3}, // 1200
		},
	}
	orderUC.Create(ctx, o)

	if o.Total != 1900 {
		t.Errorf("total = %v, want 1900", o.Total)
	}
}

func TestUnitOrderUCCreateSetsItemPricesFromProduct(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC) // price=350

	o := &entity.Order{
		CustomerID: "cust-prices",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, o)

	got, _ := orderUC.GetByID(ctx, o.ID)
	if len(got.Items) == 0 {
		t.Fatal("no items")
	}
	if got.Items[0].Price != 350 {
		t.Errorf("item price = %v, want 350", got.Items[0].Price)
	}
}

func TestUnitOrderUCGetByIDWithItems(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC)

	o := &entity.Order{
		CustomerID: "cust-get",
		Items: []entity.OrderItem{
			{ProductID: p.ID, Quantity: 2},
			{ProductID: p.ID, Quantity: 3},
		},
	}
	orderUC.Create(ctx, o)

	got, err := orderUC.GetByID(ctx, o.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.Items) != 2 {
		t.Errorf("items = %d, want 2", len(got.Items))
	}
}

func TestUnitOrderUCListByCustomerPaginationEdgeCases(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC)

	for range 8 {
		o := &entity.Order{
			CustomerID: "cust-page",
			Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
		}
		orderUC.Create(ctx, o)
	}

	tests := []struct {
		name    string
		page    int
		size    int
		wantLen int
	}{
		{"page 1", 1, 3, 3},
		{"page 2", 2, 3, 3},
		{"page 3", 3, 3, 2},
		{"past end", 10, 3, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, err := orderUC.ListByCustomer(ctx, "cust-page", tt.page, tt.size)
			if err != nil {
				t.Fatalf("ListByCustomer: %v", err)
			}
			if len(orders) != tt.wantLen {
				t.Errorf("page %d size %d = %d, want %d", tt.page, tt.size, len(orders), tt.wantLen)
			}
		})
	}
}

func TestUnitOrderUCFullLifecycleCancel(t *testing.T) {
	productUC, categoryUC, orderUC, _ := setupUC(t)
	ctx := context.Background()
	_, p := seedCatAndProduct(t, productUC, categoryUC)

	o := &entity.Order{
		CustomerID: "cust-lifecycle",
		Items:      []entity.OrderItem{{ProductID: p.ID, Quantity: 1}},
	}
	orderUC.Create(ctx, o)

	// new → process
	if err := orderUC.Process(ctx, o.ID); err != nil {
		t.Fatalf("Process: %v", err)
	}
	got, _ := orderUC.GetByID(ctx, o.ID)
	if got.Status != entity.OrderStatusProcessing {
		t.Errorf("after process: %q", got.Status)
	}

	// process → cancel
	if err := orderUC.Cancel(ctx, o.ID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	got, _ = orderUC.GetByID(ctx, o.ID)
	if got.Status != entity.OrderStatusCancelled {
		t.Errorf("after cancel: %q", got.Status)
	}
}
