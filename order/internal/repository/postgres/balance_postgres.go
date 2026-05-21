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

type BalancePostgres struct {
	pool *pgxpool.Pool
}

func NewBalancePostgres(pool *pgxpool.Pool) repository.BalanceRepository {
	return &BalancePostgres{pool: pool}
}

func (r *BalancePostgres) GetUserBalance(ctx context.Context, userID uuid.UUID) (*entity.UserBalance, error) {
	query := `
		SELECT user_id, balance, updated_at
		FROM user_balances
		WHERE user_id = $1
	`
	balance := &entity.UserBalance{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&balance.UserID, &balance.Balance, &balance.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("balance not found for user: %s", userID)
		}
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}
	return balance, nil
}

func (r *BalancePostgres) GetUserBalanceForUpdate(ctx context.Context, tx interface{}, userID uuid.UUID) (*entity.UserBalance, error) {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return nil, fmt.Errorf("invalid transaction type")
	}

	query := `
		SELECT user_id, balance, updated_at
		FROM user_balances
		WHERE user_id = $1
		FOR UPDATE
	`
	balance := &entity.UserBalance{}
	err := pgxTx.QueryRow(ctx, query, userID).Scan(
		&balance.UserID, &balance.Balance, &balance.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("balance not found for user: %s", userID)
		}
		return nil, fmt.Errorf("failed to get user balance for update: %w", err)
	}
	return balance, nil
}

func (r *BalancePostgres) UpdateBalance(ctx context.Context, tx interface{}, userID uuid.UUID, amount float64) error {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	query := `
		UPDATE user_balances
		SET balance = balance + $2, updated_at = $3
		WHERE user_id = $1
	`
	ct, err := pgxTx.Exec(ctx, query, userID, amount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("balance not found for user: %s", userID)
	}
	return nil
}

func (r *BalancePostgres) CreateTransaction(ctx context.Context, tx interface{}, txRecord *entity.Transaction) error {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("invalid transaction type")
	}

	query := `
		INSERT INTO transactions (user_id, order_id, amount, type, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	return pgxTx.QueryRow(ctx, query,
		txRecord.UserID, txRecord.OrderID, txRecord.Amount,
		txRecord.Type, txRecord.Description, time.Now(),
	).Scan(&txRecord.ID, &txRecord.CreatedAt)
}

func (r *BalancePostgres) TopUpBalance(ctx context.Context, userID uuid.UUID, amount float64, description string) (*entity.UserBalance, error) {
	query := `
		INSERT INTO user_balances (user_id, balance, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET balance = user_balances.balance + $2, updated_at = $3
		RETURNING user_id, balance, updated_at
	`
	balance := &entity.UserBalance{}
	err := r.pool.QueryRow(ctx, query, userID, amount, time.Now()).Scan(
		&balance.UserID, &balance.Balance, &balance.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to top up balance: %w", err)
	}

	// Record transaction
	txQuery := `
		INSERT INTO transactions (user_id, amount, type, description, created_at)
		VALUES ($1, $2, 'credit', $3, $4)
	`
	_, err = r.pool.Exec(ctx, txQuery, userID, amount, description, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to record top-up transaction: %w", err)
	}

	return balance, nil
}

func (r *BalancePostgres) GetTransactionHistory(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Transaction, int32, error) {
	// Count total
	var total int32
	countQuery := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count transactions: %w", err)
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, user_id, order_id, amount, type, description, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*entity.Transaction, 0)
	for rows.Next() {
		tx := &entity.Transaction{}
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.OrderID, &tx.Amount,
			&tx.Type, &tx.Description, &tx.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}
