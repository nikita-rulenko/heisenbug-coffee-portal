package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

type CategoryRepo struct {
	db *sql.DB
}

func NewCategoryRepo(db *sql.DB) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) Create(ctx context.Context, c *entity.Category) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO categories (name, slug, description, sort_order, created_at) VALUES (?, ?, ?, ?, ?)`,
		c.Name, c.Slug, c.Description, c.SortOrder, now,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = id
	c.CreatedAt = now
	return nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (*entity.Category, error) {
	c := &entity.Category{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, description, sort_order, created_at FROM categories WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.SortOrder, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	return c, err
}

func (r *CategoryRepo) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	c := &entity.Category{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, slug, description, sort_order, created_at FROM categories WHERE slug = ?`, slug,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.SortOrder, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	return c, err
}

func (r *CategoryRepo) List(ctx context.Context) ([]entity.Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, slug, description, sort_order, created_at FROM categories ORDER BY sort_order, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func (r *CategoryRepo) Update(ctx context.Context, c *entity.Category) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name=?, slug=?, description=?, sort_order=? WHERE id=?`,
		c.Name, c.Slug, c.Description, c.SortOrder, c.ID,
	)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}
