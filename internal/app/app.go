package app

import (
	grpcapp "github.com/poponyas/AuthService/internal/app/grpc"
	core_logger "github.com/poponyas/AuthService/internal/core/logger"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(
	log *core_logger.Logger,
	grpcPort int,
) *App {
	grpcApp := grpcapp.New(log, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
