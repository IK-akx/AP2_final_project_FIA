//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/config"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *pgxpool.Pool {
	cfg, err := config.Load()
	require.NoError(t, err)

	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	require.NoError(t, err)

	err = pool.Ping(context.Background())
	require.NoError(t, err)

	return pool
}

func TestBalanceRepo_TopUpAndGet(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	repo := postgres.NewBalancePostgres(pool)
	ctx := context.Background()
	userID := uuid.New()

	// Top up
	balance, err := repo.TopUpBalance(ctx, userID, 500.00, "initial deposit")
	require.NoError(t, err)
	assert.Equal(t, 500.00, balance.Balance)

	// Get balance
	balance, err = repo.GetUserBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 500.00, balance.Balance)

	// Top up again
	balance, err = repo.TopUpBalance(ctx, userID, 300.00, "second deposit")
	require.NoError(t, err)
	assert.Equal(t, 800.00, balance.Balance)
}

func TestOrderRepo_CreateAndGet(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	orderRepo := postgres.NewOrderPostgres(pool)
	ctx := context.Background()

	order := &entity.Order{
		UserID: uuid.New(),
		Status: entity.StatusConfirmed,
		Total:  250.00,
	}

	err := orderRepo.CreateOrder(ctx, order)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, order.ID)

	// Get order
	fetched, err := orderRepo.GetOrder(ctx, order.ID)
	require.NoError(t, err)
	assert.Equal(t, order.Total, fetched.Total)
	assert.Equal(t, entity.StatusConfirmed, fetched.Status)
}

func TestTransactionHistory(t *testing.T) {
	pool := setupDB(t)
	defer pool.Close()

	balanceRepo := postgres.NewBalancePostgres(pool)
	ctx := context.Background()
	userID := uuid.New()

	// Create some transactions via top-ups
	_, err := balanceRepo.TopUpBalance(ctx, userID, 100.00, "first")
	require.NoError(t, err)
	_, err = balanceRepo.TopUpBalance(ctx, userID, 200.00, "second")
	require.NoError(t, err)

	// Get history
	history, total, err := balanceRepo.GetTransactionHistory(ctx, userID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int32(2), total)
	assert.Len(t, history, 2)
}
