package grpcapp

import (
	"fmt"
	"net"

	core_logger "github.com/poponyas/AuthService/internal/core/logger"
	transport_grpc_auth "github.com/poponyas/AuthService/internal/features/auth/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	logger     *core_logger.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *core_logger.Logger,
	port int,
) *App {
	gRPCServer := grpc.NewServer()

	transport_grpc_auth.Register(gRPCServer)

	return &App{
		logger:     log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"

	reflection.Register(a.gRPCServer) // можно потои выключить(only for debugging with Postman)

	log := a.logger.With(
		zap.String("op", op),
		zap.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("starting gRPC server", zap.String("addr: ", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"
	a.logger.With(zap.String("op", op)).Info("stopping gRPC server", zap.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}
