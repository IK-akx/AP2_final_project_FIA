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
