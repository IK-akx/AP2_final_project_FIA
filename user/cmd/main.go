package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpchandler "github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/grpc"
	natsAdapter "github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/nats"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/postgres"
	redisAdapter "github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/redis"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/adapters/smtp"
	"github.com/IK-akx/AP2_final_project_FIA/user/internal/usecase"
	userpb "github.com/IK-akx/pharmacy-proto-gen/user"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log.Println("Starting user-service")

	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable")
	grpcPort := getEnv("GRPC_PORT", "50053")
	metricsPort := getEnv("METRICS_PORT", "9091")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6382")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("failed to connect postgres: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}
	log.Println("Postgres connected")

	redisCache, err := redisAdapter.NewCache(redisAddr)
	if err != nil {
		log.Printf("Redis connection warning: %v", err)
		redisCache = nil
	} else {
		log.Println("Redis connected!")
	}

	userRepo := postgres.NewUserRepo(db)
	notificationRepo := postgres.NewNotificationRepo(db)

	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("NATS connected")

	emailSender := smtp.NewEmailSender()

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Metrics server listening on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	consumer := natsAdapter.NewConsumer(nc, userRepo, notificationRepo, emailSender)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start NATS consumer: %v", err)
	}
	log.Println("NATS consumer started")

	registerUC := usecase.NewRegisterUseCase(userRepo)
	loginUC := usecase.NewLoginUseCase(userRepo)
	profileUC := usecase.NewProfileUseCase(userRepo, redisCache)

	handler := grpchandler.NewUserHandler(registerUC, loginUC, profileUC)
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	userpb.RegisterUserServiceServer(grpcServer, handler)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully")

	grpcServer.GracefulStop()
	cancel()
	time.Sleep(2 * time.Second)

	log.Println("User-service stopped")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
