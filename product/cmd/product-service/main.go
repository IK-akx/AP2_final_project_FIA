package main

import (
	"log"
	"net"
	"time"

	"github.com/fxrnweh9/product-service/config"
	grpcHandler "github.com/fxrnweh9/product-service/internal/grpc"
	natsinfra "github.com/fxrnweh9/product-service/internal/infrastructure/nats"
	repo "github.com/fxrnweh9/product-service/internal/repository/postgres"
	redisCache "github.com/fxrnweh9/product-service/internal/repository/redis"
	svc "github.com/fxrnweh9/product-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/fxrnweh9/product-service/proto/v1"
)

func main() {
	cfg := config.Load()

	db, err := repo.NewDB(cfg.PostgresDSN)
	if err != nil {
		log.Fatal("failed to init db:", err)
	}

	var ready bool
	for i := 0; i < 10; i++ {
		err = db.Conn.Ping()
		if err == nil {
			ready = true
			break
		}
		log.Println("waiting for postgres...")
		time.Sleep(2 * time.Second)
	}

	if !ready {
		log.Fatal("postgres not ready:", err)
	}

	log.Println("connecting to postgres with DSN:", cfg.PostgresDSN)
	log.Println("postgres connected successfully")
	productRepo := repo.NewProductRepository(db.Conn)

	rdb := redisCache.NewClient(cfg.RedisAddr)
	cachedRepo := redisCache.NewProductCache(rdb.Rdb, productRepo)

	natsPublisher, err := natsinfra.NewPublisher(cfg.NatsURL)
	if err != nil {
		log.Fatal(err)
	}

	productService := svc.NewProductService(cachedRepo, natsPublisher)

	lis, err := net.Listen("tcp", cfg.GRPCPort)
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	handler := grpcHandler.NewProductHandler(productService)
	pb.RegisterProductServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	log.Println("Product service started on", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
