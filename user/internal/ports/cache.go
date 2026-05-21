package ports

import "context"

type Cache interface {
	Set(ctx context.Context, key string, value []byte, ttlSeconds int) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
