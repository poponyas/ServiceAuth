package grpc_config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        int
	TokenTTL    time.Duration
	DatabaseURL string
	SigningKey  string
}

func NewConfig() (Config, error) {
	port, err := strconv.Atoi(envDefault("GRPC_PORT", "50052"))
	if err != nil || port < 1 || port > 65535 {
		return Config{}, fmt.Errorf("invalid GRPC_PORT")
	}
	ttl, err := time.ParseDuration(envDefault("TOKEN_TTL", "1h"))
	if err != nil || ttl <= 0 {
		return Config{}, fmt.Errorf("invalid TOKEN_TTL")
	}
	cfg := Config{Port: port, TokenTTL: ttl, DatabaseURL: os.Getenv("DATABASE_URL"), SigningKey: os.Getenv("JWT_SIGNING_KEY")}
	if cfg.DatabaseURL == "" || len(cfg.SigningKey) < 32 {
		return Config{}, fmt.Errorf("DATABASE_URL and JWT_SIGNING_KEY (at least 32 bytes) are required")
	}
	return cfg, nil
}

func envDefault(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
