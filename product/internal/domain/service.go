package domain

import "context"

type ProductService interface {
	CreateProduct(ctx context.Context, p *Product) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]*Product, int, error)
	CheckAvailability(ctx context.Context, productID string, qty int) (bool, int, error)
	UpdateStock(ctx context.Context, productID string, delta int) (int, int, error)
}
