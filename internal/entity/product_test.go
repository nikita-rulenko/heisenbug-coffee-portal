package entity_test

import (
	"math"
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
		{
			name:    "very small negative price",
			product: entity.Product{Name: "Тест", Price: -0.01, CategoryID: 1},
			wantErr: entity.ErrNegativePrice,
		},
		{
			name:    "large price",
			product: entity.Product{Name: "Премиум", Price: 999999.99, CategoryID: 1},
			wantErr: nil,
		},
		{
			name:    "unicode name",
			product: entity.Product{Name: "抹茶ラテ", Price: 500, CategoryID: 1},
			wantErr: nil,
		},
		{
			name:    "name with spaces only",
			product: entity.Product{Name: "   ", Price: 100, CategoryID: 1},
			wantErr: nil, // spaces are not empty
		},
		{
			name:    "emoji name",
			product: entity.Product{Name: "☕ Кофе", Price: 200, CategoryID: 1},
			wantErr: nil,
		},
		{
			name:    "very long name",
			product: entity.Product{Name: "Супер Мега Ультра Экстра Двойной Латте с Карамелью и Корицей и Ванилью и Шоколадом", Price: 500, CategoryID: 1},
			wantErr: nil,
		},
		{
			name:    "negative category large",
			product: entity.Product{Name: "Тест", Price: 100, CategoryID: -9999},
			wantErr: entity.ErrInvalidCategory,
		},
		{
			name:    "all fields valid with description",
			product: entity.Product{Name: "Латте", Price: 350, CategoryID: 1, Description: "Молочный кофе", ImageURL: "/img/latte.jpg", InStock: true},
			wantErr: nil,
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
		{"1 percent", 1, 990},
		{"99 percent", 99, 10},
		{"33.33 percent", 33.33, 666.7},
		{"tiny discount 0.01", 0.01, 999.9},
		{"exactly 100.01 returns original", 100.01, 1000},
		{"exactly -0.01 returns original", -0.01, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ApplyDiscount(tt.percent)
			if math.Abs(got-tt.want) > 0.1 {
				t.Errorf("ApplyDiscount(%v) = %v, want %v", tt.percent, got, tt.want)
			}
		})
	}
}

func TestUnitProductApplyDiscountZeroPrice(t *testing.T) {
	p := entity.Product{Price: 0}
	got := p.ApplyDiscount(50)
	if got != 0 {
		t.Errorf("ApplyDiscount(50) on zero price = %v, want 0", got)
	}
}

func TestUnitProductApplyDiscountLargePrice(t *testing.T) {
	p := entity.Product{Price: 1000000}
	got := p.ApplyDiscount(25)
	if got != 750000 {
		t.Errorf("ApplyDiscount(25) on 1M = %v, want 750000", got)
	}
}

func TestUnitProductValidatePriorityOrder(t *testing.T) {
	// empty name checked before negative price
	p := entity.Product{Name: "", Price: -10, CategoryID: 1}
	if err := p.Validate(); err != entity.ErrEmptyName {
		t.Errorf("expected ErrEmptyName first, got %v", err)
	}
}

func TestUnitProductValidateNameBeforeCategory(t *testing.T) {
	p := entity.Product{Name: "", Price: 100, CategoryID: 0}
	if err := p.Validate(); err != entity.ErrEmptyName {
		t.Errorf("expected ErrEmptyName first, got %v", err)
	}
}

func TestUnitProductValidatePriceBeforeCategory(t *testing.T) {
	p := entity.Product{Name: "Test", Price: -1, CategoryID: 0}
	if err := p.Validate(); err != entity.ErrNegativePrice {
		t.Errorf("expected ErrNegativePrice first, got %v", err)
	}
}
