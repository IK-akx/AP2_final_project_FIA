package main

import (
	"fmt"
	"os"

	"github.com/IK-akx/AP2_FINAL_PROJECT/api-gateway/config"
	"github.com/IK-akx/AP2_FINAL_PROJECT/api-gateway/internal/handlers"
	"github.com/IK-akx/AP2_FINAL_PROJECT/api-gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Initialize handlers
	orderHandler, err := handlers.NewOrderHandler(cfg.OrderServiceAddr, logger)
	if err != nil {
		logger.Fatal("failed to connect to order service", zap.Error(err))
	}

	productHandler := handlers.NewProductHandler()
	userHandler := handlers.NewUserHandler()

	// Setup Gin router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes (no auth)
	auth := r.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// Protected routes (JWT required)
	api := r.Group("/")
	api.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		// Products
		api.GET("/products", productHandler.ListProducts)
		api.GET("/products/:id", productHandler.GetProduct)

		// Orders ← ТВОЙ РАБОЧИЙ СЕРВИС
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
		api.GET("/orders/user/:userId", orderHandler.GetUserOrders)
		api.PUT("/orders/:id/cancel", orderHandler.CancelOrder)

		// Users
		api.GET("/users/profile", userHandler.GetProfile)
		api.GET("/users/:id/balance", orderHandler.GetBalance) // баланс из Order Service
	}

	// Start
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("API Gateway starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
