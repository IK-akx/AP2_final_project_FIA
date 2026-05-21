package repository

import (
	"context"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
)

// BalanceRepository defines the interface for user balance persistence
type BalanceRepository interface {
	// GetUserBalance retrieves user balance
	// Returns nil, error if user not found
	GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error)

	// GetUserBalanceForUpdate retrieves user balance with FOR UPDATE lock
	// Used within transactions for atomic operations
	GetUserBalanceForUpdate(ctx context.Context, tx interface{}, userID uuid.UUID) (*entity.UserBalance, error)

	// UpdateBalance updates user balance by adding amount (can be negative)
	// Must be called within a transaction
	UpdateBalance(ctx context.Context, tx interface{}, userID uuid.UUID, amount float64) error

	// CreateTransaction records a balance transaction
	// Must be called within a transaction
	CreateTransaction(ctx context.Context, tx interface{}, txRecord *entity.Transaction) error

	// TopUpBalance adds funds to user balance, creating user_balances record if not exists
	// Uses INSERT ... ON CONFLICT DO UPDATE
	TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64, description string) (*entity.UserBalance, error)

	// GetTransactionHistory retrieves transaction history for a user with pagination
	GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error)
}
