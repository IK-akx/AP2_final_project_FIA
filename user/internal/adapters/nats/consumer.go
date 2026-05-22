package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/domain"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/ports"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type OrderItemEvent struct {
	ProductID string  `json:"product_id"`
	Quantity  int32   `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderCreatedEvent struct {
	OrderID   string           `json:"order_id"`
	UserID    string           `json:"user_id"`
	Total     float64          `json:"total"`
	Items     []OrderItemEvent `json:"items"`
	CreatedAt string           `json:"created_at"`
}

type Consumer struct {
	nc               *nats.Conn
	userRepo         ports.UserRepository
	notificationRepo ports.NotificationRepository
	emailSender      ports.EmailSender
}

func NewConsumer(
	nc *nats.Conn,
	userRepo ports.UserRepository,
	notificationRepo ports.NotificationRepository,
	emailSender ports.EmailSender,
) *Consumer {
	return &Consumer{
		nc:               nc,
		userRepo:         userRepo,
		notificationRepo: notificationRepo,
		emailSender:      emailSender,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	sub, err := c.nc.Subscribe("order.created", func(msg *nats.Msg) {
		processCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to parse order.created: %v", err)
			return
		}

		log.Printf("Received order.created: order_id=%s, user_id=%s", event.OrderID, event.UserID)

		user, err := c.userRepo.FindByID(processCtx, event.UserID)
		if err != nil || user == nil {
			log.Printf("User not found or DB error: %s (err: %v)", event.UserID, err)
			c.logNotification(event.UserID, "order.created", "", "", "failed", "user not found or db error")
			return
		}

		subject := "Order Confirmation"
		body := fmt.Sprintf("Hello %s! Your order #%s for $%.2f has been confirmed.",
			user.FirstName, event.OrderID, event.Total)

		err = c.emailSender.Send(processCtx, user.Email, subject, body)
		if err != nil {
			log.Printf("Failed to send email: %v", err)
			c.logNotification(event.UserID, "order.created", subject, body, "failed", err.Error())
		} else {
			log.Printf("Email sent to %s", user.Email)
			c.logNotification(event.UserID, "order.created", subject, body, "sent", "")
		}
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to order.created: %w", err)
	}

	log.Printf("Successfully subscribed to order.created")

	go func() {
		<-ctx.Done()
		log.Printf("Context cancelled, unsubscribing from order.created...")
		_ = sub.Unsubscribe()
	}()

	return nil
}

func (c *Consumer) logNotification(userID, notificationType, subject, body, status, errorMsg string) {
	if c.notificationRepo == nil {
		log.Printf("notificationRepo is nil, cannot save log")
		return
	}

	logEntry := domain.NotificationLog{
		ID:       uuid.New().String(),
		UserID:   userID,
		Type:     notificationType,
		Subject:  subject,
		Body:     body,
		SentAt:   time.Now(),
		Status:   status,
		ErrorMsg: errorMsg,
	}

	if err := c.notificationRepo.Save(context.Background(), logEntry); err != nil {
		log.Printf("Failed to save notification log: %v", err)
	}
}
