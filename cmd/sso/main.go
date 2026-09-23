package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/poponyas/AuthService/internal/app"
	core_logger "github.com/poponyas/AuthService/internal/core/logger"
	grpc_config "github.com/poponyas/AuthService/internal/features/auth"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("can't load .env: %w", err)
		os.Exit(1)
	}
	cfg := grpc_config.NewConfigMust()

	logger, err := core_logger.NewLogger(
		core_logger.NewConfigMust(),
	)
	if err != nil {
		fmt.Println("can't init logger: %w", err)
		os.Exit(1)
	}

	logger.Debug("starting application...")

	application := app.New(logger, cfg.Port)

	go application.GRPCSrv.MustRun()

	// TODO: ининциализровать объект конфига

	// TODO: инициализировать логгер

	// TODO: инициализировать приложение

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.GRPCSrv.Stop()

	logger.Info("application stopped")
}
