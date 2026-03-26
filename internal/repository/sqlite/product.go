package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

type ProductRepo struct {
	db *sql.DB
}

func NewProductRepo(db *sql.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

func (r *ProductRepo) Create(ctx context.Context, p *entity.Product) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO products (category_id, name, description, price, image_url, in_stock, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		p.CategoryID, p.Name, p.Description, p.Price, p.ImageURL, p.InStock, now, now,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	p := &entity.Product{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, category_id, name, description, price, image_url, in_stock, created_at, updated_at
		 FROM products WHERE id = ?`, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.InStock, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	return p, err
}

func (r *ProductRepo) List(ctx context.Context, categoryID int64, offset, limit int) ([]entity.Product, error) {
	var rows *sql.Rows
	var err error

	if categoryID > 0 {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, category_id, name, description, price, image_url, in_stock, created_at, updated_at
			 FROM products WHERE category_id = ? ORDER BY name LIMIT ? OFFSET ?`,
			categoryID, limit, offset,
		)
	} else {
		rows, err = r.db.QueryContext(ctx,
			`SELECT id, category_id, name, description, price, image_url, in_stock, created_at, updated_at
			 FROM products ORDER BY name LIMIT ? OFFSET ?`,
			limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (r *ProductRepo) Update(ctx context.Context, p *entity.Product) error {
	p.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE products SET category_id=?, name=?, description=?, price=?, image_url=?, in_stock=?, updated_at=?
		 WHERE id=?`,
		p.CategoryID, p.Name, p.Description, p.Price, p.ImageURL, p.InStock, p.UpdatedAt, p.ID,
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

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM products WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *ProductRepo) Search(ctx context.Context, query string, limit int) ([]entity.Product, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, category_id, name, description, price, image_url, in_stock, created_at, updated_at
		 FROM products WHERE name LIKE ? OR description LIKE ? ORDER BY name LIMIT ?`,
		"%"+query+"%", "%"+query+"%", limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (r *ProductRepo) Count(ctx context.Context, categoryID int64) (int, error) {
	var count int
	var err error
	if categoryID > 0 {
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE category_id = ?`, categoryID).Scan(&count)
	} else {
		err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&count)
	}
	return count, err
}

func scanProducts(rows *sql.Rows) ([]entity.Product, error) {
	var products []entity.Product
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.Price, &p.ImageURL, &p.InStock, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}
