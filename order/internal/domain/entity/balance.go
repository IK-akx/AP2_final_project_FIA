package entity

import (
	"time"

	"github.com/google/uuid"
)

// Transaction type constants
const (
	TransactionTypeDebit  = "debit"
	TransactionTypeCredit = "credit"
)

// UserBalance represents a user's account balance
type UserBalance struct {
	UserID    uuid.UUID
	Balance   float64
	UpdatedAt time.Time
}

// Transaction represents a balance transaction (debit or credit)
type Transaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	OrderID     *uuid.UUID // nullable
	Amount      float64
	Type        string
	Description string
	CreatedAt   time.Time
}
