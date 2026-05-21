package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	JWTSecret          string
	OrderServiceAddr   string
	ProductServiceAddr string
	UserServiceAddr    string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-key-change-in-production"),
		OrderServiceAddr:   getEnv("ORDER_SERVICE_ADDR", "localhost:50052"),
		ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "localhost:50051"),
		UserServiceAddr:    getEnv("USER_SERVICE_ADDR", "localhost:50053"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
