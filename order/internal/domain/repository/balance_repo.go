package repository

import (
	"context"

	"github.com/IKakx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
)

// BalanceRepository defines the interface for user balance persistence
type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error)
	UpdateBalance(ctx context.Context, userID uuid.UUID, amount float64, transactionType string, orderID *uuid.UUID, description string) error
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error)
	TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64, description string) (*entity.UserBalance, error)
}
