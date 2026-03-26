package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

type OrderRepo struct {
	db *sql.DB
}

func NewOrderRepo(db *sql.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

func (r *OrderRepo) Create(ctx context.Context, o *entity.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	result, err := tx.ExecContext(ctx,
		`INSERT INTO orders (customer_id, status, total, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		o.CustomerID, o.Status, o.Total, now, now,
	)
	if err != nil {
		return err
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return err
	}

	for i := range o.Items {
		item := &o.Items[i]
		res, err := tx.ExecContext(ctx,
			`INSERT INTO order_items (order_id, product_id, quantity, price) VALUES (?, ?, ?, ?)`,
			orderID, item.ProductID, item.Quantity, item.Price,
		)
		if err != nil {
			return err
		}
		itemID, _ := res.LastInsertId()
		item.ID = itemID
		item.OrderID = orderID
	}

	o.ID = orderID
	o.CreatedAt = now
	o.UpdatedAt = now

	return tx.Commit()
}

func (r *OrderRepo) GetByID(ctx context.Context, id int64) (*entity.Order, error) {
	o := &entity.Order{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, customer_id, status, total, created_at, updated_at FROM orders WHERE id = ?`, id,
	).Scan(&o.ID, &o.CustomerID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	items, err := r.GetItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	return o, nil
}

func (r *OrderRepo) ListByCustomer(ctx context.Context, customerID string, offset, limit int) ([]entity.Order, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, customer_id, status, total, created_at, updated_at
		 FROM orders WHERE customer_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		customerID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []entity.Order
	for rows.Next() {
		var o entity.Order
		if err := rows.Scan(&o.ID, &o.CustomerID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, id int64, status entity.OrderStatus) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status=?, updated_at=? WHERE id=?`,
		status, time.Now(), id,
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

func (r *OrderRepo) GetItems(ctx context.Context, orderID int64) ([]entity.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, product_id, quantity, price FROM order_items WHERE order_id = ?`, orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.OrderItem
	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
