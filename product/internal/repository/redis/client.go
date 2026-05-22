package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	Rdb *redis.Client
}

func NewClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &Client{Rdb: rdb}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Rdb.Ping(ctx).Err()
}
