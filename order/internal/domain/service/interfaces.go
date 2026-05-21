package service

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
)

// ProductServiceClient defines the interface for Product Service gRPC calls
type ProductServiceClient interface {
	CheckAvailability(ctx context.Context, productID uuid.UUID, quantity int32) (available bool, currentStock int32, err error)
	UpdateStock(ctx context.Context, productID uuid.UUID, delta int32) error
}

// NATSPublisher defines the interface for NATS event publishing
type NATSPublisher interface {
	PublishOrderCreated(ctx context.Context, order *entity.Order) error
}

// OrderService defines the business logic interface for orders
type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, items []CreateOrderItem) (*entity.Order, error)
	GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error)
	CancelOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)
}

// BalanceService defines the business logic interface for user balance
type BalanceService interface {
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error)
	TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64) (*entity.UserBalance, error)
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error)
}

// CreateOrderItem represents an item in a new order request
type CreateOrderItem struct {
	ProductID uuid.UUID
	Quantity  int32
}
