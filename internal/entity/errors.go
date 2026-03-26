package entity

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrEmptyName       = errors.New("name cannot be empty")
	ErrEmptySlug       = errors.New("slug cannot be empty")
	ErrEmptyContent    = errors.New("content cannot be empty")
	ErrNegativePrice   = errors.New("price cannot be negative")
	ErrInvalidCategory = errors.New("invalid category ID")
	ErrInvalidProduct  = errors.New("invalid product ID")
	ErrInvalidQuantity = errors.New("quantity must be positive")
	ErrEmptyCustomerID = errors.New("customer ID cannot be empty")
	ErrEmptyOrder      = errors.New("order must have at least one item")
	ErrAlreadyExists   = errors.New("already exists")
	ErrInvalidStatus   = errors.New("invalid order status transition")
)
