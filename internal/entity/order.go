package entity

import "time"

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "new"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID         int64       `json:"id"`
	CustomerID string      `json:"customer_id"`
	Status     OrderStatus `json:"status"`
	Items      []OrderItem `json:"items"`
	Total      float64     `json:"total"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return ErrEmptyCustomerID
	}
	if len(o.Items) == 0 {
		return ErrEmptyOrder
	}
	for _, item := range o.Items {
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
		if item.ProductID <= 0 {
			return ErrInvalidProduct
		}
	}
	return nil
}

func (o *Order) CalculateTotal() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.Total = total
	return total
}

func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusNew || o.Status == OrderStatusProcessing
}

func (o *Order) CanComplete() bool {
	return o.Status == OrderStatusProcessing
}
