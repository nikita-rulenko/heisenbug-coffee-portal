package usecase

import (
	"context"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/repository"
)

type NewsUseCase struct {
	repo repository.NewsRepository
}

func NewNewsUseCase(repo repository.NewsRepository) *NewsUseCase {
	return &NewsUseCase{repo: repo}
}

func (uc *NewsUseCase) Create(ctx context.Context, n *entity.NewsItem) error {
	if err := n.Validate(); err != nil {
		return err
	}
	return uc.repo.Create(ctx, n)
}

func (uc *NewsUseCase) GetByID(ctx context.Context, id int64) (*entity.NewsItem, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *NewsUseCase) List(ctx context.Context, page, pageSize int) ([]entity.NewsItem, int, error) {
	if pageSize <= 0 {
		pageSize = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize

	items, err := uc.repo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := uc.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (uc *NewsUseCase) Update(ctx context.Context, n *entity.NewsItem) error {
	if err := n.Validate(); err != nil {
		return err
	}
	return uc.repo.Update(ctx, n)
}

func (uc *NewsUseCase) Delete(ctx context.Context, id int64) error {
	return uc.repo.Delete(ctx, id)
}
