package repository

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
)

// OrderRepository defines the interface for order persistence
type OrderRepository interface {
	CreateOrder(ctx context.Context, order *entity.Order) error
	GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string) error
}
