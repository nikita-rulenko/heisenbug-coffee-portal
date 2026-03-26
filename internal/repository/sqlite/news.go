package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

type NewsRepo struct {
	db *sql.DB
}

func NewNewsRepo(db *sql.DB) *NewsRepo {
	return &NewsRepo{db: db}
}

func (r *NewsRepo) Create(ctx context.Context, n *entity.NewsItem) error {
	now := time.Now()
	if n.PublishedAt.IsZero() {
		n.PublishedAt = now
	}
	result, err := r.db.ExecContext(ctx,
		`INSERT INTO news (title, content, author, published_at, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		n.Title, n.Content, n.Author, n.PublishedAt, now, now,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	n.ID = id
	n.CreatedAt = now
	n.UpdatedAt = now
	return nil
}

func (r *NewsRepo) GetByID(ctx context.Context, id int64) (*entity.NewsItem, error) {
	n := &entity.NewsItem{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, content, author, published_at, created_at, updated_at FROM news WHERE id = ?`, id,
	).Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	return n, err
}

func (r *NewsRepo) List(ctx context.Context, offset, limit int) ([]entity.NewsItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, content, author, published_at, created_at, updated_at
		 FROM news ORDER BY published_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.NewsItem
	for rows.Next() {
		var n entity.NewsItem
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Author, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, rows.Err()
}

func (r *NewsRepo) Update(ctx context.Context, n *entity.NewsItem) error {
	n.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx,
		`UPDATE news SET title=?, content=?, author=?, published_at=?, updated_at=? WHERE id=?`,
		n.Title, n.Content, n.Author, n.PublishedAt, n.UpdatedAt, n.ID,
	)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *NewsRepo) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM news WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return entity.ErrNotFound
	}
	return nil
}

func (r *NewsRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM news`).Scan(&count)
	return count, err
}
