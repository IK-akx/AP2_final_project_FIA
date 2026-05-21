package service

import (
	"context"
	"fmt"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/repository"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type BalanceSvc struct {
	balanceRepo repository.BalanceRepository
	logger      *zap.Logger
}

func NewBalanceService(balanceRepo repository.BalanceRepository, logger *zap.Logger) BalanceService {
	return &BalanceSvc{
		balanceRepo: balanceRepo,
		logger:      logger,
	}
}

func (s *BalanceSvc) GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error) {
	return s.balanceRepo.GetUserBalance(ctx, userID)
}

func (s *BalanceSvc) TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64) (*entity.UserBalance, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("top-up amount must be positive: %.2f", amount)
	}

	description := fmt.Sprintf("Balance top-up: %.2f", amount)
	balance, err := s.balanceRepo.TopUpBalance(ctx, userID, amount, description)
	if err != nil {
		return nil, err
	}

	s.logger.Info("balance topped up",
		zap.String("user_id", userID.String()),
		zap.Float64("amount", amount),
		zap.Float64("new_balance", balance.Balance),
	)

	return balance, nil
}

func (s *BalanceSvc) GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.balanceRepo.GetTransactionHistory(ctx, userID, page, limit)
}
