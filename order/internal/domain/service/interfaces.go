package service

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
)

// ProductServiceClient defines the interface for Product Service gRPC calls
type ProductServiceClient interface {
	// CheckAvailability checks if product has enough stock
	CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int32) (available bool, currentStock int32, err error)

	// UpdateStock updates product stock (negative = decrease, positive = increase)
	UpdateStock(ctx context.Context, productID uuid.UUID, delta int32) error
	GetProductPrice(ctx context.Context, productID uuid.UUID) (float64, error)
}

// NATSPublisher defines the interface for NATS event publishing
type NATSPublisher interface {
	// PublishOrderCreated publishes order.created event
	PublishOrderCreated(ctx context.Context, order *entity.Order) error
}

// OrderService defines the business logic interface for orders
type OrderService interface {
	// CreateOrder validates availability, checks balance, creates order atomically
	CreateOrder(ctx context.Context, userID uuid.UUID, items []CreateOrderItem) (*entity.Order, error)

	// GetOrder retrieves an order with its items (from cache if available)
	GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)

	// GetUserOrders retrieves user's orders with pagination
	GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error)

	// CancelOrder cancels a confirmed order and refunds money
	CancelOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)
}

// BalanceService defines the business logic interface for user balance
type BalanceService interface {
	// GetUserBalance retrieves user balance
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error)

	// TopUpBalance adds funds to user balance
	TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64) (*entity.UserBalance, error)

	// InitBalance
	InitBalance(ctx context.Context, userID uuid.UUID) error

	// GetTransactionHistory retrieves transaction history for a user
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error)
}

// CreateOrderItem represents an item in a new order request
type CreateOrderItem struct {
	ProductID uuid.UUID
	Quantity  int32
}
