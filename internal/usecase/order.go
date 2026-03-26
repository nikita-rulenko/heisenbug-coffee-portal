package usecase

import (
	"context"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/repository"
)

type OrderUseCase struct {
	repo        repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewOrderUseCase(repo repository.OrderRepository, productRepo repository.ProductRepository) *OrderUseCase {
	return &OrderUseCase{repo: repo, productRepo: productRepo}
}

func (uc *OrderUseCase) Create(ctx context.Context, o *entity.Order) error {
	if err := o.Validate(); err != nil {
		return err
	}

	for i := range o.Items {
		product, err := uc.productRepo.GetByID(ctx, o.Items[i].ProductID)
		if err != nil {
			return entity.ErrInvalidProduct
		}
		o.Items[i].Price = product.Price
	}

	o.CalculateTotal()
	o.Status = entity.OrderStatusNew

	return uc.repo.Create(ctx, o)
}

func (uc *OrderUseCase) GetByID(ctx context.Context, id int64) (*entity.Order, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *OrderUseCase) ListByCustomer(ctx context.Context, customerID string, page, pageSize int) ([]entity.Order, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return uc.repo.ListByCustomer(ctx, customerID, offset, pageSize)
}

func (uc *OrderUseCase) Cancel(ctx context.Context, id int64) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !order.CanCancel() {
		return entity.ErrInvalidStatus
	}
	return uc.repo.UpdateStatus(ctx, id, entity.OrderStatusCancelled)
}

func (uc *OrderUseCase) Process(ctx context.Context, id int64) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order.Status != entity.OrderStatusNew {
		return entity.ErrInvalidStatus
	}
	return uc.repo.UpdateStatus(ctx, id, entity.OrderStatusProcessing)
}

func (uc *OrderUseCase) Complete(ctx context.Context, id int64) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !order.CanComplete() {
		return entity.ErrInvalidStatus
	}
	return uc.repo.UpdateStatus(ctx, id, entity.OrderStatusCompleted)
}
