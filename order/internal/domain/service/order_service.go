package service

import (
	"context"
	"fmt"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/repository"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/repository/redis"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type OrderSvc struct {
	orderRepo   repository.OrderRepository
	balanceRepo repository.BalanceRepository
	cache       *redis.OrderCache
	productSvc  ProductServiceClient
	natsPub     NATSPublisher
	pool        *pgxpool.Pool
	logger      *zap.Logger
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	balanceRepo repository.BalanceRepository,
	cache *redis.OrderCache,
	productSvc ProductServiceClient,
	natsPub NATSPublisher,
	pool *pgxpool.Pool,
	logger *zap.Logger,
) OrderService {
	return &OrderSvc{
		orderRepo:   orderRepo,
		balanceRepo: balanceRepo,
		cache:       cache,
		productSvc:  productSvc,
		natsPub:     natsPub,
		pool:        pool,
		logger:      logger,
	}
}

// CreateOrder implements the full atomic order creation flow
func (s *OrderSvc) CreateOrder(ctx context.Context, userID uuid.UUID, items []CreateOrderItem) (*entity.Order, error) {
	// Step 1: Check product availability for each item
	var total float64
	orderItems := make([]*entity.OrderItem, len(items))

	for i, item := range items {
		available, currentStock, err := s.productSvc.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to check availability for product %s: %w", item.ProductID, err)
		}
		if !available {
			return nil, fmt.Errorf("product %s: insufficient stock (requested: %d, available: %d)",
				item.ProductID, item.Quantity, currentStock)
		}

		price, err := s.productSvc.GetProductPrice(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for product %s: %w", item.ProductID, err)
		}

		total += price * float64(item.Quantity)
		orderItems[i] = &entity.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     price,
		}
	}

	// Step 2-9: Atomic PostgreSQL transaction
	order := &entity.Order{
		UserID: userID,
		Status: entity.StatusConfirmed,
		Total:  total,
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // Rollback если не закоммитили (no-op после Commit)

	// Step 3: Lock and check balance
	balance, err := s.balanceRepo.GetUserBalanceForUpdate(ctx, tx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	if balance.Balance < total {
		return nil, fmt.Errorf("insufficient balance: have %.2f, need %.2f", balance.Balance, total)
	}

	// Step 4: Create order
	if err := s.createOrderTx(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Step 5: Create order items
	for _, item := range orderItems {
		item.OrderID = order.ID
	}
	if err := s.createOrderItemsTx(ctx, tx, orderItems); err != nil {
		return nil, fmt.Errorf("failed to create order items: %w", err)
	}

	// Step 6: Deduct balance
	if err := s.balanceRepo.UpdateBalance(ctx, tx, userID, -total); err != nil {
		return nil, fmt.Errorf("failed to deduct balance: %w", err)
	}

	// Step 7: Record transaction
	txRecord := &entity.Transaction{
		UserID:      userID,
		OrderID:     &order.ID,
		Amount:      total,
		Type:        entity.TransactionTypeDebit,
		Description: fmt.Sprintf("Payment for order %s", order.ID.String()),
	}
	if err := s.balanceRepo.CreateTransaction(ctx, tx, txRecord); err != nil {
		return nil, fmt.Errorf("failed to record transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Step 10 (after commit): Update stock in Product Service
	for _, item := range orderItems {
		delta := -item.Quantity // negative = decrease stock
		if err := s.productSvc.UpdateStock(ctx, item.ProductID, delta); err != nil {
			s.logger.Error("failed to update stock after order",
				zap.String("order_id", order.ID.String()),
				zap.String("product_id", item.ProductID.String()),
				zap.Error(err),
			)
		}
	}

	// Step 11 (after commit): Publish NATS event
	order.Items = make([]entity.OrderItem, len(orderItems))
	for i, item := range orderItems {
		order.Items[i] = *item
	}
	if err := s.natsPub.PublishOrderCreated(ctx, order); err != nil {
		s.logger.Error("failed to publish order.created event",
			zap.String("order_id", order.ID.String()),
			zap.Error(err),
		)
	}

	if err := s.cache.InvalidateUserOrders(ctx, userID); err != nil {
		s.logger.Warn("failed to invalidate cache", zap.Error(err))
	}

	return order, nil
}

func (s *OrderSvc) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	// Cache-Aside: check Redis first
	order, err := s.cache.GetOrder(ctx, id)
	if err != nil {
		s.logger.Warn("cache get failed, falling back to DB", zap.Error(err))
	}
	if order != nil {
		return order, nil
	}

	// Cache miss — get from PostgreSQL
	order, err = s.orderRepo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load items
	items, err := s.orderRepo.GetOrderItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	// Конвертируем []*entity.OrderItem → []entity.OrderItem
	order.Items = make([]entity.OrderItem, len(items))
	for i, item := range items {
		order.Items[i] = *item
	}

	// Store in cache
	if err := s.cache.SetOrder(ctx, order); err != nil {
		s.logger.Warn("failed to cache order", zap.Error(err))
	}

	return order, nil
}

func (s *OrderSvc) GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int32) ([]*entity.Order, int32, error) {
	// Cache-Aside for first page only (most common case)
	if page == 1 {
		orders, err := s.cache.GetUserOrders(ctx, userID, page)
		if err != nil {
			s.logger.Warn("cache get failed, falling back to DB", zap.Error(err))
		}
		if orders != nil {
			_, total, err := s.orderRepo.GetUserOrders(ctx, userID, page, limit)
			return orders, total, err
		}
	}

	orders, total, err := s.orderRepo.GetUserOrders(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Cache first page
	if page == 1 {
		if err := s.cache.SetUserOrders(ctx, userID, page, orders); err != nil {
			s.logger.Warn("failed to cache user orders", zap.Error(err))
		}
	}

	return orders, total, nil
}

func (s *OrderSvc) CancelOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	// Get the order first
	order, err := s.orderRepo.GetOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.CanBeCancelled() {
		return nil, fmt.Errorf("order %s cannot be cancelled (current status: %s)", id, order.Status)
	}

	// Atomic cancel with refund
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update order status
	if err := s.updateOrderStatusTx(ctx, tx, id, entity.StatusCancelled); err != nil {
		return nil, err
	}

	// Refund balance
	if err := s.balanceRepo.UpdateBalance(ctx, tx, order.UserID, order.Total); err != nil {
		return nil, fmt.Errorf("failed to refund balance: %w", err)
	}

	// Record refund transaction
	refundTx := &entity.Transaction{
		UserID:      order.UserID,
		OrderID:     &order.ID,
		Amount:      order.Total,
		Type:        entity.TransactionTypeCredit,
		Description: fmt.Sprintf("Refund for cancelled order %s", order.ID.String()),
	}
	if err := s.balanceRepo.CreateTransaction(ctx, tx, refundTx); err != nil {
		return nil, fmt.Errorf("failed to record refund transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After commit: restore stock
	items, err := s.orderRepo.GetOrderItems(ctx, id)
	if err != nil {
		s.logger.Error("failed to get order items for stock restore", zap.Error(err))
	} else {
		for _, item := range items {
			if err := s.productSvc.UpdateStock(ctx, item.ProductID, item.Quantity); err != nil {
				s.logger.Error("failed to restore stock after cancel",
					zap.String("order_id", id.String()),
					zap.String("product_id", item.ProductID.String()),
					zap.Error(err),
				)
			}
		}
	}

	// Invalidate cache
	if err := s.cache.DeleteOrder(ctx, id); err != nil {
		s.logger.Warn("failed to invalidate order cache", zap.Error(err))
	}
	if err := s.cache.InvalidateUserOrders(ctx, order.UserID); err != nil {
		s.logger.Warn("failed to invalidate user orders cache", zap.Error(err))
	}

	order.Status = entity.StatusCancelled
	return order, nil
}

// === Transaction helper methods ===

func (s *OrderSvc) createOrderTx(ctx context.Context, tx pgx.Tx, order *entity.Order) error {
	query := `
		INSERT INTO orders (user_id, status, total, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return tx.QueryRow(ctx, query, order.UserID, order.Status, order.Total).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

func (s *OrderSvc) createOrderItemsTx(ctx context.Context, tx pgx.Tx, items []*entity.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	for _, item := range items {
		if err := tx.QueryRow(ctx, query, item.OrderID, item.ProductID, item.Quantity, item.Price).
			Scan(&item.ID); err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}
	return nil
}

func (s *OrderSvc) updateOrderStatusTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, status string) error {
	query := `UPDATE orders SET status = $2, updated_at = NOW() WHERE id = $1`
	ct, err := tx.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %s", id)
	}
	return nil
}
