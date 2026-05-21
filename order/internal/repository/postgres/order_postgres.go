package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderPostgres struct {
	pool *pgxpool.Pool
}

func NewOrderPostgres(pool *pgxpool.Pool) repository.OrderRepository {
	return &OrderPostgres{pool: pool}
}

func (r *OrderPostgres) CreateOrder(ctx context.Context, order *entity.Order) error {
	query := `
		INSERT INTO orders (user_id, status, total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		order.UserID, order.Status, order.Total, time.Now(), time.Now(),
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

// CreateOrderTx creates an order within an existing transaction
func (r *OrderPostgres) CreateOrderTx(ctx context.Context, tx pgx.Tx, order *entity.Order) error {
	query := `
		INSERT INTO orders (user_id, status, total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return tx.QueryRow(ctx, query,
		order.UserID, order.Status, order.Total, time.Now(), time.Now(),
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

func (r *OrderPostgres) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	query := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	order := &entity.Order{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.Status, &order.Total,
		&order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	return order, nil
}

func (r *OrderPostgres) GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error) {
	// Count total
	var total int32
	countQuery := `SELECT COUNT(*) FROM orders WHERE user_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	// Get paginated orders
	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	orders := make([]*entity.Order, 0)
	for rows.Next() {
		order := &entity.Order{}
		if err := rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.Total,
			&order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

func (r *OrderPostgres) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	ct, err := r.pool.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %s", id)
	}
	return nil
}

// UpdateOrderStatusTx updates order status within a transaction
func (r *OrderPostgres) UpdateOrderStatusTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	ct, err := tx.Exec(ctx, query, id, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %s", id)
	}
	return nil
}

func (r *OrderPostgres) CreateOrderItems(ctx context.Context, items []*entity.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	for _, item := range items {
		if err := r.pool.QueryRow(ctx, query,
			item.OrderID, item.ProductID, item.Quantity, item.Price,
		).Scan(&item.ID); err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}
	return nil
}

// CreateOrderItemsTx creates order items within a transaction
func (r *OrderPostgres) CreateOrderItemsTx(ctx context.Context, tx pgx.Tx, items []*entity.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	for _, item := range items {
		if err := tx.QueryRow(ctx, query,
			item.OrderID, item.ProductID, item.Quantity, item.Price,
		).Scan(&item.ID); err != nil {
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}
	return nil
}

func (r *OrderPostgres) GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	items := make([]*entity.OrderItem, 0)
	for rows.Next() {
		item := &entity.OrderItem{}
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}
