package postgres

import (
	"context"
	"database/sql"

	"github.com/fxrnweh9/product-service/internal/domain"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func nullableUUID(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	query := `
	INSERT INTO products (
		id, name, description, category_id,
		price, stock, manufacturer, requires_rx,
		image_url, created_at, updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, nullableUUID(p.CategoryID),
		p.Price, p.Stock, p.Manufacturer, p.RequiresRX,
		p.ImageURL, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
	SELECT id, name, description, COALESCE(category_id::text, ''),
		   price, stock, manufacturer, requires_rx,
		   image_url, created_at, updated_at
	FROM products
	WHERE id = $1
	`
	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.CategoryID,
		&p.Price, &p.Stock, &p.Manufacturer, &p.RequiresRX,
		&p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) List(ctx context.Context, limit, offset int) ([]*domain.Product, int, error) {
	query := `
	SELECT id, name, description, COALESCE(category_id::text, ''),
		   price, stock, manufacturer, requires_rx,
		   image_url, created_at, updated_at
	FROM products
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.CategoryID,
			&p.Price, &p.Stock, &p.Manufacturer, &p.RequiresRX,
			&p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, &p)
	}

	var total int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM products").Scan(&total)

	return products, total, nil
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id string, newStock int) (int, error) {
	var old int
	err := r.db.QueryRowContext(ctx,
		"SELECT stock FROM products WHERE id = $1", id,
	).Scan(&old)
	if err != nil {
		return 0, err
	}

	_, err = r.db.ExecContext(ctx,
		"UPDATE products SET stock = $1, updated_at = NOW() WHERE id = $2",
		newStock, id,
	)
	if err != nil {
		return 0, err
	}
	return old, nil
}
