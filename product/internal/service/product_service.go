package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/fxrnweh9/product-service/internal/domain"
	"github.com/google/uuid"
)

type EventPublisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

type ProductService struct {
	repo      domain.ProductRepository
	publisher EventPublisher
}

func NewProductService(repo domain.ProductRepository, publisher EventPublisher) *ProductService {
	return &ProductService{
		repo:      repo,
		publisher: publisher,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	p.ID = uuid.New().String()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, int, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *ProductService) CheckAvailability(ctx context.Context, productID string, qty int) (bool, int, error) {
	p, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		return false, 0, err
	}

	if p.Stock >= qty {
		return true, p.Stock, nil
	}

	return false, p.Stock, nil
}

func (s *ProductService) UpdateStock(ctx context.Context, productID string, delta int) (int, int, error) {
	p, err := s.repo.GetByID(ctx, productID)
	if err != nil {
		return 0, 0, err
	}

	newStock := p.Stock + delta
	if newStock < 0 {
		newStock = 0
	}

	old, err := s.repo.UpdateStock(ctx, productID, newStock)
	if err != nil {
		return 0, 0, err
	}

	// NATS event
	event := struct {
		ProductID string    `json:"product_id"`
		NewStock  int       `json:"new_stock"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		ProductID: productID,
		NewStock:  newStock,
		UpdatedAt: time.Now(),
	}

	data, _ := json.Marshal(event)

	_ = s.publisher.Publish(ctx, "stock.updated", data)

	return old, newStock, nil
}
