package repository

import (
	"context"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
	List(ctx context.Context, categoryID int64, offset, limit int) ([]entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id int64) error
	Search(ctx context.Context, query string, limit int) ([]entity.Product, error)
	Count(ctx context.Context, categoryID int64) (int, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id int64) (*entity.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Category, error)
	List(ctx context.Context) ([]entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id int64) error
}

type NewsRepository interface {
	Create(ctx context.Context, item *entity.NewsItem) error
	GetByID(ctx context.Context, id int64) (*entity.NewsItem, error)
	List(ctx context.Context, offset, limit int) ([]entity.NewsItem, error)
	Update(ctx context.Context, item *entity.NewsItem) error
	Delete(ctx context.Context, id int64) error
	Count(ctx context.Context) (int, error)
}

type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id int64) (*entity.Order, error)
	ListByCustomer(ctx context.Context, customerID string, offset, limit int) ([]entity.Order, error)
	UpdateStatus(ctx context.Context, id int64, status entity.OrderStatus) error
	GetItems(ctx context.Context, orderID int64) ([]entity.OrderItem, error)
}
