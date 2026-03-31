package entity_test

import (
	"math"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestUnitProductApplyDiscountVariousPrices(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		want  float64
	}{
		{"price 0.01", 0.01, 0.0075},
		{"price 0.99", 0.99, 0.7425},
		{"price 1", 1, 0.75},
		{"price 10", 10, 7.5},
		{"price 50", 50, 37.5},
		{"price 99.99", 99.99, 74.9925},
		{"price 100", 100, 75},
		{"price 500", 500, 375},
		{"price 999.99", 999.99, 749.9925},
		{"price 1000", 1000, 750},
		{"price 5000", 5000, 3750},
		{"price 10000", 10000, 7500},
		{"price 50000", 50000, 37500},
		{"price 100000", 100000, 75000},
		{"price 999999", 999999, 749999.25},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := entity.Product{Price: tt.price}
			got := p.ApplyDiscount(25)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("ApplyDiscount(25) on price %v = %v, want %v", tt.price, got, tt.want)
			}
		})
	}
}

func TestUnitProductApplyDiscountBoundaryPercents(t *testing.T) {
	p := entity.Product{Price: 1000}
	tests := []struct {
		name    string
		percent float64
		want    float64
	}{
		{"0%", 0, 1000},
		{"0.001%", 0.001, 999.99},
		{"0.01%", 0.01, 999.9},
		{"0.1%", 0.1, 999},
		{"1%", 1, 990},
		{"50%", 50, 500},
		{"99%", 99, 10},
		{"99.9%", 99.9, 1},
		{"99.99%", 99.99, 0.1},
		{"100%", 100, 0},
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

func TestUnitProductValidateFieldCombinations(t *testing.T) {
	tests := []struct {
		name    string
		product entity.Product
		wantErr error
	}{
		{"valid_valid_valid", entity.Product{Name: "Кофе", Price: 100, CategoryID: 1}, nil},
		{"empty_valid_valid", entity.Product{Name: "", Price: 100, CategoryID: 1}, entity.ErrEmptyName},
		{"valid_neg_valid", entity.Product{Name: "Кофе", Price: -1, CategoryID: 1}, entity.ErrNegativePrice},
		{"valid_valid_zero", entity.Product{Name: "Кофе", Price: 100, CategoryID: 0}, entity.ErrInvalidCategory},
		{"empty_neg_valid", entity.Product{Name: "", Price: -1, CategoryID: 1}, entity.ErrEmptyName},
		{"empty_valid_zero", entity.Product{Name: "", Price: 100, CategoryID: 0}, entity.ErrEmptyName},
		{"valid_neg_zero", entity.Product{Name: "Кофе", Price: -1, CategoryID: 0}, entity.ErrNegativePrice},
		{"empty_neg_zero", entity.Product{Name: "", Price: -1, CategoryID: 0}, entity.ErrEmptyName},
		{"valid_zero_valid", entity.Product{Name: "Кофе", Price: 0, CategoryID: 1}, nil},
		{"valid_zero_zero", entity.Product{Name: "Кофе", Price: 0, CategoryID: 0}, entity.ErrInvalidCategory},
		{"empty_zero_valid", entity.Product{Name: "", Price: 0, CategoryID: 1}, entity.ErrEmptyName},
		{"empty_zero_zero", entity.Product{Name: "", Price: 0, CategoryID: 0}, entity.ErrEmptyName},
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

func TestUnitProductApplyDiscountConsistency(t *testing.T) {
	p := entity.Product{Price: 1000}
	first := p.ApplyDiscount(30)
	second := p.ApplyDiscount(30)
	if first != second {
		t.Errorf("ApplyDiscount not consistent: %v != %v", first, second)
	}
}

func TestUnitProductApplyDiscountVerySmallPrice(t *testing.T) {
	p := entity.Product{Price: 0.001}
	got := p.ApplyDiscount(50)
	if math.Abs(got-0.0005) > 0.0001 {
		t.Errorf("ApplyDiscount(50) on 0.001 = %v, want ~0.0005", got)
	}
}

func TestUnitProductValidateDescriptionOptional(t *testing.T) {
	p := entity.Product{Name: "Кофе", Price: 100, CategoryID: 1}
	if err := p.Validate(); err != nil {
		t.Errorf("product without description should be valid: %v", err)
	}
}

func TestUnitProductValidateImageURLOptional(t *testing.T) {
	p := entity.Product{Name: "Кофе", Price: 100, CategoryID: 1, Description: "Описание"}
	if err := p.Validate(); err != nil {
		t.Errorf("product without imageURL should be valid: %v", err)
	}
}

func TestUnitProductValidateInStockDefault(t *testing.T) {
	p := entity.Product{Name: "Кофе", Price: 100, CategoryID: 1}
	if p.InStock {
		t.Error("default InStock should be false")
	}
	if err := p.Validate(); err != nil {
		t.Errorf("product with InStock=false should be valid: %v", err)
	}
}
