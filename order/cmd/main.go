package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IK-akx/AP2_FINAL_PROJECT/order/config"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/delivery/grpc"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/domain/service"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/repository/postgres"
	rediscache "github.com/IK-akx/AP2_FINAL_PROJECT/order/internal/repository/redis"
	"github.com/IK-akx/AP2_FINAL_PROJECT/order/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init("info"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	log := logger.Log
	defer logger.Sync()

	log.Info("starting order service")

	// Connect to PostgreSQL
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatal("failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("failed to ping PostgreSQL", zap.Error(err))
	}
	log.Info("connected to PostgreSQL")

	// Connect to Redis
	redisClient := goredis.NewClient(&goredis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPass,
		DB:       cfg.RedisDB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("failed to connect to Redis", zap.Error(err))
	}
	log.Info("connected to Redis")

	// Connect to NATS
	natsPub, err := service.NewNATSPublisher(cfg.NatsURL, log)
	if err != nil {
		log.Fatal("failed to connect to NATS", zap.Error(err))
	}
	defer natsPub.Close()
	log.Info("connected to NATS")

	// gRPC client for Product Service
	productClient, err := service.NewProductClient(cfg.ProductServiceAddr, log)
	if err != nil {
		log.Fatal("failed to create product service client", zap.Error(err))
	}
	log.Info("product service client created")

	// Initialize repositories
	orderRepo := postgres.NewOrderPostgres(pool)
	balanceRepo := postgres.NewBalancePostgres(pool)
	orderCache := rediscache.NewOrderCache(redisClient)

	// Initialize services
	orderSvc := service.NewOrderService(orderRepo, balanceRepo, orderCache, productClient, natsPub, pool, log)
	balanceSvc := service.NewBalanceService(balanceRepo, log)

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%s", cfg.GRPCPort)
	grpcServer, err := grpc.NewServer(grpcAddr, orderSvc, balanceSvc, log)
	if err != nil {
		log.Fatal("failed to create gRPC server", zap.Error(err))
	}

	grpcServer.Start()
	log.Info("order service started", zap.String("grpc_addr", grpcAddr))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down order service")
	grpcServer.Stop()
	log.Info("order service stopped")
}
