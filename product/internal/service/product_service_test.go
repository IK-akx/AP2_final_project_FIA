package service_test

import (
	"context"
	"testing"

	"github.com/fxrnweh9/product-service/internal/domain"
	"github.com/fxrnweh9/product-service/internal/service"
)

type fakeRepo struct {
	products map[string]*domain.Product
}

func (f *fakeRepo) Create(ctx context.Context, p *domain.Product) error {
	f.products[p.ID] = p
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return f.products[id], nil
}

func (f *fakeRepo) List(ctx context.Context, limit, offset int) ([]*domain.Product, int, error) {
	var res []*domain.Product
	for _, p := range f.products {
		res = append(res, p)
	}
	return res, len(res), nil
}

func (f *fakeRepo) UpdateStock(ctx context.Context, id string, newStock int) (int, error) {
	old := f.products[id].Stock
	f.products[id].Stock = newStock
	return old, nil
}

func TestCheckAvailability(t *testing.T) {
	repo := &fakeRepo{
		products: map[string]*domain.Product{
			"1": {ID: "1", Stock: 10},
		},
	}

	svc := service.NewProductService(repo, nil)

	ok, stock, err := svc.CheckAvailability(context.Background(), "1", 5)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("expected available")
	}

	if stock != 10 {
		t.Fatalf("expected 10 got %d", stock)
	}
}
