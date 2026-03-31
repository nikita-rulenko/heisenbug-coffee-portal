package entity_test

import (
	"math"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitOrderCalculateTotalVariousItemCounts(t *testing.T) {
	tests := []struct {
		name  string
		items []entity.OrderItem
		want  float64
	}{
		{"1 item", []entity.OrderItem{{Price: 100, Quantity: 1}}, 100},
		{"2 items", []entity.OrderItem{{Price: 100, Quantity: 1}, {Price: 200, Quantity: 2}}, 500},
		{"3 items", []entity.OrderItem{{Price: 100, Quantity: 1}, {Price: 200, Quantity: 1}, {Price: 300, Quantity: 1}}, 600},
		{"5 items same price", []entity.OrderItem{{Price: 50, Quantity: 1}, {Price: 50, Quantity: 1}, {Price: 50, Quantity: 1}, {Price: 50, Quantity: 1}, {Price: 50, Quantity: 1}}, 250},
		{"all qty 1", []entity.OrderItem{{Price: 10, Quantity: 1}, {Price: 20, Quantity: 1}, {Price: 30, Quantity: 1}}, 60},
		{"mixed quantities", []entity.OrderItem{{Price: 100, Quantity: 3}, {Price: 50, Quantity: 10}}, 800},
		{"single expensive", []entity.OrderItem{{Price: 99999, Quantity: 1}}, 99999},
		{"many cheap", []entity.OrderItem{{Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}, {Price: 1, Quantity: 1}}, 10},
		{"fractional prices", []entity.OrderItem{{Price: 0.5, Quantity: 3}, {Price: 1.5, Quantity: 2}}, 4.5},
		{"large qty", []entity.OrderItem{{Price: 10, Quantity: 1000}}, 10000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := entity.Order{Items: tt.items}
			got := o.CalculateTotal()
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("CalculateTotal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnitOrderValidateMultipleInvalidItems(t *testing.T) {
	tests := []struct {
		name    string
		items   []entity.OrderItem
		wantErr error
	}{
		{"first item bad qty", []entity.OrderItem{{ProductID: 1, Quantity: 0}, {ProductID: 2, Quantity: 1}}, entity.ErrInvalidQuantity},
		{"last item bad qty", []entity.OrderItem{{ProductID: 1, Quantity: 1}, {ProductID: 2, Quantity: 0}}, entity.ErrInvalidQuantity},
		{"middle item bad qty", []entity.OrderItem{{ProductID: 1, Quantity: 1}, {ProductID: 2, Quantity: -1}, {ProductID: 3, Quantity: 1}}, entity.ErrInvalidQuantity},
		{"first item bad product", []entity.OrderItem{{ProductID: 0, Quantity: 1}, {ProductID: 2, Quantity: 1}}, entity.ErrInvalidProduct},
		{"last item bad product", []entity.OrderItem{{ProductID: 1, Quantity: 1}, {ProductID: 0, Quantity: 1}}, entity.ErrInvalidProduct},
		{"all items bad qty", []entity.OrderItem{{ProductID: 1, Quantity: 0}, {ProductID: 2, Quantity: 0}}, entity.ErrInvalidQuantity},
		{"5th of 10 bad", []entity.OrderItem{{ProductID: 1, Quantity: 1}, {ProductID: 2, Quantity: 1}, {ProductID: 3, Quantity: 1}, {ProductID: 4, Quantity: 1}, {ProductID: 0, Quantity: 1}, {ProductID: 6, Quantity: 1}, {ProductID: 7, Quantity: 1}, {ProductID: 8, Quantity: 1}, {ProductID: 9, Quantity: 1}, {ProductID: 10, Quantity: 1}}, entity.ErrInvalidProduct},
		{"qty before product", []entity.OrderItem{{ProductID: 0, Quantity: 0}}, entity.ErrInvalidQuantity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := entity.Order{CustomerID: "cust-1", Items: tt.items}
			err := o.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnitOrderStatusTransitionMatrix(t *testing.T) {
	tests := []struct {
		name       string
		status     entity.OrderStatus
		canCancel  bool
		canComplete bool
	}{
		{"new", entity.OrderStatusNew, true, false},
		{"processing", entity.OrderStatusProcessing, true, true},
		{"completed", entity.OrderStatusCompleted, false, false},
		{"cancelled", entity.OrderStatusCancelled, false, false},
		{"unknown", entity.OrderStatus("unknown"), false, false},
		{"empty string", entity.OrderStatus(""), false, false},
		{"random string", entity.OrderStatus("shipped"), false, false},
		{"uppercase NEW", entity.OrderStatus("NEW"), false, false},
		{"mixed case New", entity.OrderStatus("New"), false, false},
		{"with space", entity.OrderStatus(" new"), false, false},
		{"partial", entity.OrderStatus("proc"), false, false},
		{"unicode status", entity.OrderStatus("новый"), false, false},
		{"null-like", entity.OrderStatus("null"), false, false},
		{"pending", entity.OrderStatus("pending"), false, false},
		{"refunded", entity.OrderStatus("refunded"), false, false},
		{"deleted", entity.OrderStatus("deleted"), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := entity.Order{Status: tt.status}
			if got := o.CanCancel(); got != tt.canCancel {
				t.Errorf("CanCancel() = %v, want %v", got, tt.canCancel)
			}
			if got := o.CanComplete(); got != tt.canComplete {
				t.Errorf("CanComplete() = %v, want %v", got, tt.canComplete)
			}
		})
	}
}

func TestUnitOrderCalculateTotalNilItems(t *testing.T) {
	o := entity.Order{Items: nil}
	got := o.CalculateTotal()
	if got != 0 {
		t.Errorf("CalculateTotal() on nil items = %v, want 0", got)
	}
}

func TestUnitOrderCalculateTotalMixedPrices(t *testing.T) {
	o := entity.Order{
		Items: []entity.OrderItem{
			{Price: 0.01, Quantity: 1},
			{Price: 99999, Quantity: 1},
		},
	}
	got := o.CalculateTotal()
	if math.Abs(got-99999.01) > 0.1 {
		t.Errorf("CalculateTotal() = %v, want ~99999.01", got)
	}
}

func TestUnitOrderCalculateTotalOverwritesPrevious(t *testing.T) {
	o := entity.Order{
		Total: 999,
		Items: []entity.OrderItem{{Price: 50, Quantity: 2}},
	}
	got := o.CalculateTotal()
	if got != 100 {
		t.Errorf("CalculateTotal() = %v, want 100", got)
	}
	if o.Total != 100 {
		t.Errorf("Total field = %v, want 100", o.Total)
	}
}

func TestUnitOrderValidateItemsValidationOrder(t *testing.T) {
	// quantity checked before productID in each item
	o := entity.Order{
		CustomerID: "cust-1",
		Items:      []entity.OrderItem{{ProductID: 0, Quantity: 0}},
	}
	err := o.Validate()
	if err != entity.ErrInvalidQuantity {
		t.Errorf("expected ErrInvalidQuantity (checked first), got %v", err)
	}
}

func TestUnitOrderValidateSingleItemAllFields(t *testing.T) {
	o := entity.Order{
		CustomerID: "customer-42",
		Items: []entity.OrderItem{
			{ProductID: 1, Quantity: 5, Price: 350, OrderID: 10, ID: 100},
		},
	}
	if err := o.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}
