package entity

import "time"

type Product struct {
	ID          int64     `json:"id"`
	CategoryID  int64     `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"`
	InStock     bool      `json:"in_stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Product) Validate() error {
	if p.Name == "" {
		return ErrEmptyName
	}
	if p.Price < 0 {
		return ErrNegativePrice
	}
	if p.CategoryID <= 0 {
		return ErrInvalidCategory
	}
	return nil
}

func (p *Product) ApplyDiscount(percent float64) float64 {
	if percent < 0 || percent > 100 {
		return p.Price
	}
	return p.Price * (1 - percent/100)
}
