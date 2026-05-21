package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const subjectOrderCreated = "order.created"

// OrderCreatedEvent is the NATS message payload for order.created
type OrderCreatedEvent struct {
	OrderID   string           `json:"order_id"`
	UserID    string           `json:"user_id"`
	Total     float64          `json:"total"`
	Items     []OrderItemEvent `json:"items"`
	UserEmail string           `json:"user_email"` // empty, User Service resolves
	CreatedAt string           `json:"created_at"`
}

// OrderItemEvent represents an item in the NATS event
type OrderItemEvent struct {
	ProductID string  `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float64 `json:"price"`
}

// NATSPub implements NATSPublisher interface
type NATSPub struct {
	conn   *nats.Conn
	logger *zap.Logger
}

func NewNATSPublisher(url string, logger *zap.Logger) (*NATSPub, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &NATSPub{
		conn:   conn,
		logger: logger,
	}, nil
}

func (p *NATSPub) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	items := make([]OrderItemEvent, len(order.Items))
	for i, item := range order.Items {
		items[i] = OrderItemEvent{
			ProductID: item.ProductID.String(),
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	event := OrderCreatedEvent{
		OrderID:   order.ID.String(),
		UserID:    order.UserID.String(),
		Total:     order.Total,
		Items:     items,
		UserEmail: "", // User Service resolves this
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal order event: %w", err)
	}

	if err := p.conn.Publish(subjectOrderCreated, data); err != nil {
		p.logger.Error("failed to publish order.created event",
			zap.String("order_id", order.ID.String()),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish event: %w", err)
	}

	p.logger.Info("published order.created event",
		zap.String("order_id", order.ID.String()),
		zap.String("subject", subjectOrderCreated),
	)

	return nil
}

func (p *NATSPub) Close() {
	p.conn.Close()
}
