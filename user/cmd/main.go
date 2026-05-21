package main

import (
	"context"
	"log"
	"os"

	"github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/postgres"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/usecase"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Starting user-service")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable")

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}
	log.Println("Postgres connected")

	userRepo := postgres.NewUserRepo(db)
	registerUC := usecase.NewRegisterUseCase(userRepo)
	loginUC := usecase.NewLoginUseCase(userRepo)

	_ = registerUC
	_ = loginUC

	select {}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
