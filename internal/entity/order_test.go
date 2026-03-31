package entity_test

import (
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitOrderValidate(t *testing.T) {
	tests := []struct {
		name    string
		order   entity.Order
		wantErr error
	}{
		{
			name: "valid order",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{{ProductID: 1, Quantity: 2}},
			},
			wantErr: nil,
		},
		{
			name: "empty customer ID",
			order: entity.Order{
				CustomerID: "",
				Items:      []entity.OrderItem{{ProductID: 1, Quantity: 1}},
			},
			wantErr: entity.ErrEmptyCustomerID,
		},
		{
			name: "empty items",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{},
			},
			wantErr: entity.ErrEmptyOrder,
		},
		{
			name: "zero quantity",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{{ProductID: 1, Quantity: 0}},
			},
			wantErr: entity.ErrInvalidQuantity,
		},
		{
			name: "negative quantity",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{{ProductID: 1, Quantity: -3}},
			},
			wantErr: entity.ErrInvalidQuantity,
		},
		{
			name: "invalid product ID",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{{ProductID: 0, Quantity: 1}},
			},
			wantErr: entity.ErrInvalidProduct,
		},
		{
			name: "negative product ID",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{{ProductID: -5, Quantity: 1}},
			},
			wantErr: entity.ErrInvalidProduct,
		},
		{
			name: "nil items",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      nil,
			},
			wantErr: entity.ErrEmptyOrder,
		},
		{
			name: "multiple valid items",
			order: entity.Order{
				CustomerID: "cust-1",
				Items: []entity.OrderItem{
					{ProductID: 1, Quantity: 2},
					{ProductID: 2, Quantity: 3},
					{ProductID: 3, Quantity: 1},
				},
			},
			wantErr: nil,
		},
		{
			name: "second item invalid quantity",
			order: entity.Order{
				CustomerID: "cust-1",
				Items: []entity.OrderItem{
					{ProductID: 1, Quantity: 2},
					{ProductID: 2, Quantity: 0},
				},
			},
			wantErr: entity.ErrInvalidQuantity,
		},
		{
			name: "second item invalid product",
			order: entity.Order{
				CustomerID: "cust-1",
				Items: []entity.OrderItem{
					{ProductID: 1, Quantity: 1},
					{ProductID: 0, Quantity: 1},
				},
			},
			wantErr: entity.ErrInvalidProduct,
		},
		{
			name: "large quantity",
			order: entity.Order{
				CustomerID: "cust-1",
				Items:      []entity.OrderItem{{ProductID: 1, Quantity: 10000}},
			},
			wantErr: nil,
		},
		{
			name: "unicode customer ID",
			order: entity.Order{
				CustomerID: "клиент-42",
				Items:      []entity.OrderItem{{ProductID: 1, Quantity: 1}},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.order.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnitOrderCalculateTotal(t *testing.T) {
	o := entity.Order{
		Items: []entity.OrderItem{
			{Price: 350, Quantity: 2},
			{Price: 250, Quantity: 1},
			{Price: 400, Quantity: 3},
		},
	}

	total := o.CalculateTotal()
	expected := 350*2 + 250*1 + 400*3.0
	if total != expected {
		t.Errorf("CalculateTotal() = %v, want %v", total, expected)
	}
	if o.Total != expected {
		t.Errorf("order.Total = %v, want %v", o.Total, expected)
	}
}

func TestUnitOrderCalculateTotalEmpty(t *testing.T) {
	o := entity.Order{Items: []entity.OrderItem{}}
	total := o.CalculateTotal()
	if total != 0 {
		t.Errorf("CalculateTotal() for empty = %v, want 0", total)
	}
}

func TestUnitOrderCalculateTotalSingleItem(t *testing.T) {
	o := entity.Order{
		Items: []entity.OrderItem{{Price: 500, Quantity: 1}},
	}
	if got := o.CalculateTotal(); got != 500 {
		t.Errorf("CalculateTotal() = %v, want 500", got)
	}
}

func TestUnitOrderCalculateTotalLargeOrder(t *testing.T) {
	items := make([]entity.OrderItem, 50)
	for i := range items {
		items[i] = entity.OrderItem{Price: 100, Quantity: 2}
	}
	o := entity.Order{Items: items}
	if got := o.CalculateTotal(); got != 10000 {
		t.Errorf("CalculateTotal() = %v, want 10000", got)
	}
}

func TestUnitOrderCalculateTotalFloatPrecision(t *testing.T) {
	o := entity.Order{
		Items: []entity.OrderItem{
			{Price: 0.1, Quantity: 10},
			{Price: 0.2, Quantity: 5},
		},
	}
	got := o.CalculateTotal()
	if got < 1.9 || got > 2.1 {
		t.Errorf("CalculateTotal() = %v, want ~2.0", got)
	}
}

func TestUnitOrderCanCancel(t *testing.T) {
	tests := []struct {
		status entity.OrderStatus
		want   bool
	}{
		{entity.OrderStatusNew, true},
		{entity.OrderStatusProcessing, true},
		{entity.OrderStatusCompleted, false},
		{entity.OrderStatusCancelled, false},
	}
	for _, tt := range tests {
		o := entity.Order{Status: tt.status}
		if got := o.CanCancel(); got != tt.want {
			t.Errorf("CanCancel() with status %q = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestUnitOrderCanComplete(t *testing.T) {
	tests := []struct {
		status entity.OrderStatus
		want   bool
	}{
		{entity.OrderStatusNew, false},
		{entity.OrderStatusProcessing, true},
		{entity.OrderStatusCompleted, false},
		{entity.OrderStatusCancelled, false},
	}
	for _, tt := range tests {
		o := entity.Order{Status: tt.status}
		if got := o.CanComplete(); got != tt.want {
			t.Errorf("CanComplete() with status %q = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestUnitOrderCanCancelUnknownStatus(t *testing.T) {
	o := entity.Order{Status: "unknown"}
	if o.CanCancel() {
		t.Error("CanCancel() with unknown status should be false")
	}
}

func TestUnitOrderCanCompleteUnknownStatus(t *testing.T) {
	o := entity.Order{Status: "unknown"}
	if o.CanComplete() {
		t.Error("CanComplete() with unknown status should be false")
	}
}

func TestUnitOrderValidatePriorityCustomerFirst(t *testing.T) {
	o := entity.Order{
		CustomerID: "",
		Items:      nil,
	}
	if err := o.Validate(); err != entity.ErrEmptyCustomerID {
		t.Errorf("expected ErrEmptyCustomerID first, got %v", err)
	}
}

func TestUnitOrderCalculateTotalSetsField(t *testing.T) {
	o := entity.Order{
		Total: 999, // pre-existing value
		Items: []entity.OrderItem{{Price: 100, Quantity: 1}},
	}
	got := o.CalculateTotal()
	if got != 100 {
		t.Errorf("CalculateTotal() = %v, want 100", got)
	}
	if o.Total != 100 {
		t.Errorf("Total field not updated: %v", o.Total)
	}
}
