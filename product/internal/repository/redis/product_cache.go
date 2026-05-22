package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fxrnweh9/product-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ProductCache struct {
	rdb  *redis.Client
	next domain.ProductRepository
}

func NewProductCache(rdb *redis.Client, next domain.ProductRepository) *ProductCache {
	return &ProductCache{
		rdb:  rdb,
		next: next,
	}
}

func (c *ProductCache) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	key := fmt.Sprintf("product:%s", id)

	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var p domain.Product
		if json.Unmarshal([]byte(val), &p) == nil {
			return &p, nil
		}
	}

	p, err := c.next.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	data, _ := json.Marshal(p)
	c.rdb.Set(ctx, key, data, 5*time.Minute)

	return p, nil
}

func (c *ProductCache) List(ctx context.Context, limit, offset int) ([]*domain.Product, int, error) {
	key := fmt.Sprintf("products:list:%d:%d", limit, offset)

	val, err := c.rdb.Get(ctx, key).Result()
	if err == nil {
		var cached struct {
			Products []*domain.Product
			Total    int
		}

		if json.Unmarshal([]byte(val), &cached) == nil {
			return cached.Products, cached.Total, nil
		}
	}

	products, total, err := c.next.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	data, _ := json.Marshal(struct {
		Products []*domain.Product
		Total    int
	}{products, total})

	c.rdb.Set(ctx, key, data, 2*time.Minute)

	return products, total, nil
}

func (c *ProductCache) UpdateStock(ctx context.Context, id string, newStock int) (int, error) {
	old, err := c.next.UpdateStock(ctx, id, newStock)
	if err != nil {
		return 0, err
	}

	c.rdb.Del(ctx, fmt.Sprintf("product:%s", id))
	c.rdb.FlushDB(ctx)

	return old, nil
}

func (p *ProductCache) Create(ctx context.Context, product *domain.Product) error {
	return p.next.Create(ctx, product)
}
