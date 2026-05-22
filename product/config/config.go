package config

import (
	"os"
)

type Config struct {
	PostgresDSN string
	RedisAddr   string
	NatsURL     string
	GRPCPort    string
}

func Load() *Config {
	return &Config{
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5433/product_db?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6380"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
		GRPCPort:    getEnv("GRPC_PORT", ":50051"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
