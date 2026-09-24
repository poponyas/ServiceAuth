package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	core_logger "github.com/poponyas/AuthService/internal/core/logger"
	"github.com/poponyas/AuthService/internal/core/secrets"
	grpc_config "github.com/poponyas/AuthService/internal/features/auth"
	service_auth "github.com/poponyas/AuthService/internal/features/auth/service/auth"
	service_storage "github.com/poponyas/AuthService/internal/features/auth/service/storage"
	transport "github.com/poponyas/AuthService/internal/features/auth/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	_ = godotenv.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if os.Getenv("VAULT_ADDR") != "" {
		data, err := secrets.Load(ctx, "service-auth")
		if err != nil {
			return err
		}
		if v := data["database_url"]; v != "" {
			_ = os.Setenv("DATABASE_URL", v)
		}
		if v := data["jwt_signing_key"]; v != "" {
			_ = os.Setenv("JWT_SIGNING_KEY", v)
		}
	}
	cfg, err := grpc_config.NewConfig()
	if err != nil {
		return err
	}
	folder := os.Getenv("LOG_FOLDER")
	if folder == "" {
		folder = "out/logs"
	}
	log, err := core_logger.NewLogger(core_logger.Config{Level: "INFO", Folder: folder})
	if err != nil {
		return err
	}
	defer log.Close()
	dbctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	db, err := pgxpool.New(dbctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(dbctx); err != nil {
		return err
	}
	store := &service_storage.Storage{DB: db}
	svc := service_auth.New(log, store, store, store, cfg.TokenTTL, []byte(cfg.SigningKey))
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	transport.Register(server, svc, []byte(cfg.SigningKey))
	reflection.Register(server)
	go func() { <-ctx.Done(); server.GracefulStop() }()
	return server.Serve(listener)
}
