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
