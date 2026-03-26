package entity_test

import (
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitProductValidate(t *testing.T) {
	tests := []struct {
		name    string
		product entity.Product
		wantErr error
	}{
		{
			name:    "valid product",
			product: entity.Product{Name: "Латте", Price: 350, CategoryID: 1},
			wantErr: nil,
		},
		{
			name:    "empty name",
			product: entity.Product{Name: "", Price: 350, CategoryID: 1},
			wantErr: entity.ErrEmptyName,
		},
		{
			name:    "negative price",
			product: entity.Product{Name: "Латте", Price: -10, CategoryID: 1},
			wantErr: entity.ErrNegativePrice,
		},
		{
			name:    "zero price is valid",
			product: entity.Product{Name: "Бесплатный образец", Price: 0, CategoryID: 1},
			wantErr: nil,
		},
		{
			name:    "zero category ID",
			product: entity.Product{Name: "Латте", Price: 350, CategoryID: 0},
			wantErr: entity.ErrInvalidCategory,
		},
		{
			name:    "negative category ID",
			product: entity.Product{Name: "Латте", Price: 350, CategoryID: -1},
			wantErr: entity.ErrInvalidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.product.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnitProductApplyDiscount(t *testing.T) {
	p := entity.Product{Price: 1000}

	tests := []struct {
		name    string
		percent float64
		want    float64
	}{
		{"10 percent", 10, 900},
		{"50 percent", 50, 500},
		{"0 percent", 0, 1000},
		{"100 percent", 100, 0},
		{"negative percent returns original", -5, 1000},
		{"over 100 percent returns original", 150, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ApplyDiscount(tt.percent)
			if got != tt.want {
				t.Errorf("ApplyDiscount(%v) = %v, want %v", tt.percent, got, tt.want)
			}
		})
	}
}
