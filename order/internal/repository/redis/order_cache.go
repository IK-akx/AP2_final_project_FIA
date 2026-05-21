package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type OrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewOrderCache(client *redis.Client) *OrderCache {
	return &OrderCache{
		client: client,
		ttl:    3 * time.Minute,
	}
}

func (c *OrderCache) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	key := fmt.Sprintf("order:%s", id.String())
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, fmt.Errorf("failed to get order from cache: %w", err)
	}

	var order entity.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached order: %w", err)
	}

	return &order, nil
}

func (c *OrderCache) SetOrder(ctx context.Context, order *entity.Order) error {
	key := fmt.Sprintf("order:%s", order.ID.String())
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order for cache: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *OrderCache) DeleteOrder(ctx context.Context, id uuid.UUID) error {
	key := fmt.Sprintf("order:%s", id.String())
	return c.client.Del(ctx, key).Err()
}

func (c *OrderCache) GetUserOrders(ctx context.Context, userID uuid.UUID, page int32) ([]*entity.Order, error) {
	key := fmt.Sprintf("orders:user:%s:%d", userID.String(), page)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, fmt.Errorf("failed to get user orders from cache: %w", err)
	}

	var orders []*entity.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached orders: %w", err)
	}

	return orders, nil
}

func (c *OrderCache) SetUserOrders(ctx context.Context, userID uuid.UUID, page int32, orders []*entity.Order) error {
	key := fmt.Sprintf("orders:user:%s:%d", userID.String(), page)
	data, err := json.Marshal(orders)
	if err != nil {
		return fmt.Errorf("failed to marshal orders for cache: %w", err)
	}

	return c.client.Set(ctx, key, data, 1*time.Minute).Err()
}

func (c *OrderCache) InvalidateUserOrders(ctx context.Context, userID uuid.UUID) error {
	// Delete all cached pages for this user
	pattern := fmt.Sprintf("orders:user:%s:*", userID.String())
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete cache key %s: %w", iter.Val(), err)
		}
	}
	return iter.Err()
}
