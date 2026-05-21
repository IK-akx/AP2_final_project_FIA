package main

import (
	"log"
	"os"
)

func main() {
	log.Println("Start user-service")

	grpcPort := getEnv("GRPC_PORT", "50053")
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6382")
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	jwtSecret := getEnv("JWT_SECRET", "secret")

	log.Printf("gRPC port:     %s", grpcPort)
	log.Printf("DB URL:        %s", dbURL)
	log.Printf("Redis addr:    %s", redisAddr)
	log.Printf("NATS URL:      %s", natsURL)
	log.Printf("JWT secret ok: %v", jwtSecret != "")

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
