package postgres_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/fxrnweh9/product-service/internal/domain"
	"github.com/fxrnweh9/product-service/internal/repository/postgres"
)

func setupDB(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "user",
			"POSTGRES_PASSWORD": "pass",
			"POSTGRES_DB":       "pharmacy",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := "postgres://user:pass@" + host + ":" + port.Port() + "/pharmacy?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			description TEXT,
			created_at TIMESTAMP DEFAULT now()
		)`,
		`CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category_id UUID REFERENCES categories(id),
			price DECIMAL(10,2) NOT NULL DEFAULT 0,
			stock INT NOT NULL DEFAULT 0,
			manufacturer VARCHAR(255),
			requires_rx BOOLEAN DEFAULT false,
			image_url VARCHAR(500),
			created_at TIMESTAMP DEFAULT now(),
			updated_at TIMESTAMP DEFAULT now()
		)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			t.Fatal("migration error:", err)
		}
	}

	return db, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func TestProductRepository_CreateAndGet(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	repo := postgres.NewProductRepository(db)
	ctx := context.Background()

	p := &domain.Product{
		ID:         uuid.New().String(),
		Name:       "Paracetamol",
		Price:      2.50,
		Stock:      50,
		CategoryID: "", // пусто — репозиторий должен передать NULL
	}

	err := repo.Create(ctx, p)
	if err != nil {
		t.Fatal("create error:", err)
	}

	got, err := repo.GetByID(ctx, p.ID)
	if err != nil {
		t.Fatal("get error:", err)
	}

	if got.Name != "Paracetamol" {
		t.Fatalf("expected Paracetamol got %s", got.Name)
	}
}
