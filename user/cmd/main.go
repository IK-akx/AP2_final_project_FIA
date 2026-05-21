package main

import (
	"context"
	"log"
	"net"
	"os"

	grpchandler "github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/grpc"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/postgres"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/usecase"
	userpb "github.com/IK-akx/pharmacy-proto-gen/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Starting user-service...")

	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable")
	grpcPort := getEnv("GRPC_PORT", "50053")

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}
	log.Println("Postgres connected")

	userRepo := postgres.NewUserRepo(db)
	registerUC := usecase.NewRegisterUseCase(userRepo)
	loginUC := usecase.NewLoginUseCase(userRepo)
	profileUC := usecase.NewProfileUseCase(userRepo)

	handler := grpchandler.NewUserHandler(registerUC, loginUC, profileUC)

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on :%s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
