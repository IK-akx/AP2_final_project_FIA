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
		PostgresDSN: "host=postgres port=5432 user=user password=pass dbname=pharmacy sslmode=disable",
		RedisAddr:   "pharmacy-redis:6379",
		NatsURL:     "nats://pharmacy-nats:4222",
		GRPCPort:    getEnv("GRPC_PORT", ":50051"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
