package entity

import (
	"time"

	"github.com/google/uuid"
)

// Order status constants
const (
	StatusPending   = "pending"
	StatusConfirmed = "confirmed"
	StatusCancelled = "cancelled"
	StatusDelivered = "delivered"
)

// Order represents a customer order
type Order struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Status    string
	Total     float64
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderItem represents a single item within an order
type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int32
	Price     float64
}

// OrderStats represents aggregated order statistics
type OrderStats struct {
	TotalOrders    int32
	TotalRevenue   float64
	OrdersByStatus map[string]int32
}

// IsValidTransition checks if the status transition is allowed
func (o *Order) IsValidTransition(newStatus string) bool {
	validTransitions := map[string][]string{
		StatusPending:   {StatusConfirmed, StatusCancelled},
		StatusConfirmed: {StatusCancelled, StatusDelivered},
		StatusCancelled: {},
		StatusDelivered: {},
	}

	allowed, exists := validTransitions[o.Status]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

// CanBeCancelled returns true if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == StatusConfirmed
}
