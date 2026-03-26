package usecase

import (
	"context"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/repository"
)

type ProductUseCase struct {
	repo     repository.ProductRepository
	catRepo  repository.CategoryRepository
}

func NewProductUseCase(repo repository.ProductRepository, catRepo repository.CategoryRepository) *ProductUseCase {
	return &ProductUseCase{repo: repo, catRepo: catRepo}
}

func (uc *ProductUseCase) Create(ctx context.Context, p *entity.Product) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, err := uc.catRepo.GetByID(ctx, p.CategoryID); err != nil {
		return entity.ErrInvalidCategory
	}
	return uc.repo.Create(ctx, p)
}

func (uc *ProductUseCase) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *ProductUseCase) List(ctx context.Context, categoryID int64, page, pageSize int) ([]entity.Product, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	products, err := uc.repo.List(ctx, categoryID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.repo.Count(ctx, categoryID)
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (uc *ProductUseCase) Update(ctx context.Context, p *entity.Product) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, p)
}

func (uc *ProductUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}

func (uc *ProductUseCase) Search(ctx context.Context, query string, limit int) ([]entity.Product, error) {
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.Search(ctx, query, limit)
}
