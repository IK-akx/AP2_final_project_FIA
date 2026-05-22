package domain

import "context"

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, limit, offset int) ([]*Product, int, error)
	UpdateStock(ctx context.Context, id string, newStock int) (old int, err error)
}

type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	GetByID(ctx context.Context, id string) (*Category, error)
}
