package main

import (
	"fmt"
	"os"
	"time"

	"github.com/IK-akx/AP2_FINAL_PROJECT/api-gateway/config"
	"github.com/IK-akx/AP2_FINAL_PROJECT/api-gateway/internal/handlers"
	"github.com/IK-akx/AP2_FINAL_PROJECT/api-gateway/internal/middleware"
	"github.com/gin-contrib/cors"
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

	orderHandler, err := handlers.NewOrderHandler(cfg.OrderServiceAddr, logger)
	if err != nil {
		logger.Fatal("failed to connect to order service", zap.Error(err))
	}

	userHandler, err := handlers.NewUserHandler(cfg.UserServiceAddr, logger)
	if err != nil {
		logger.Fatal("failed to connect to user service", zap.Error(err))
	}

	productHandler, err := handlers.NewProductHandler(cfg.ProductServiceAddr, logger)
	if err != nil {
		logger.Fatal("failed to connect to product service", zap.Error(err))
	}

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// Protected routes
	api := r.Group("/")
	api.Use(middleware.JWTAuth(cfg.UserServiceAddr))
	{
		api.GET("/products", productHandler.ListProducts)
		api.GET("/products/:id", productHandler.GetProduct)

		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders/:id", orderHandler.GetOrder)
		api.GET("/orders/user/:userId", orderHandler.GetUserOrders)
		api.PUT("/orders/:id/cancel", orderHandler.CancelOrder)

		api.GET("/users/profile", userHandler.GetProfile)
		api.PUT("/users/profile", userHandler.UpdateProfile)
		api.GET("/users/:id/balance", orderHandler.GetBalance)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("API Gateway starting", zap.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
