package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the service
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisHost string
	RedisPort string
	RedisPass string
	RedisDB   int

	// NATS
	NatsURL string

	// gRPC
	GRPCPort string

	// Product Service gRPC client
	ProductServiceAddr string

	// Observability
	JaegerEndpoint string
	PrometheusPort string
}

// Load reads configuration from .env file and environment variables
func Load() (*Config, error) {
	// Load .env file if exists (silent fail if not found)
	_ = godotenv.Load()

	cfg := &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5434"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "order_db"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6381"),
		RedisPass: getEnv("REDIS_PASSWORD", ""),
		RedisDB:   getEnvAsInt("REDIS_DB", 0),

		NatsURL: getEnv("NATS_URL", "nats://localhost:4222"),

		GRPCPort: getEnv("GRPC_PORT", "50052"),

		ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "localhost:50051"),

		JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
		PrometheusPort: getEnv("PROMETHEUS_PORT", "9102"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN returns PostgreSQL connection string
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// RedisAddr returns Redis address
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func (c *Config) validate() error {
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.NatsURL == "" {
		return fmt.Errorf("NATS_URL is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		var intVal int
		fmt.Sscanf(value, "%d", &intVal)
		return intVal
	}
	return defaultValue
}
