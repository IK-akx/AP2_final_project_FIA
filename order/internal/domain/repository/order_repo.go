package repository

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	// CreateOrder inserts a new order (without items) into the database
	// Returns the created order with generated ID and timestamps
	CreateOrder(ctx context.Context, order *entity.Order) error

	// GetOrder retrieves an order by ID
	// Does NOT load order items — use GetOrderItems for that
	GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)

	// GetUserOrders retrieves orders for a user with pagination
	// Returns orders (without items), total count
	GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error)

	// UpdateOrderStatus changes the status of an order
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error

	// CreateOrderItems inserts order items in batch
	CreateOrderItems(ctx context.Context, items []*entity.OrderItem) error

	// GetOrderItems retrieves all items for an order
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]*entity.OrderItem, error)
}
