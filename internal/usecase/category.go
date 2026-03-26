package usecase

import (
	"context"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/repository"
)

type CategoryUseCase struct {
	repo repository.CategoryRepository
}

func NewCategoryUseCase(repo repository.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

func (uc *CategoryUseCase) Create(ctx context.Context, c *entity.Category) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, c)
}

func (uc *CategoryUseCase) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *CategoryUseCase) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	return uc.repo.GetBySlug(ctx, slug)
}

func (uc *CategoryUseCase) List(ctx context.Context) ([]entity.Category, error) {
	return uc.repo.List(ctx)
}

func (uc *CategoryUseCase) Update(ctx context.Context, c *entity.Category) error {
	if err := c.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, c)
}

func (uc *CategoryUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
