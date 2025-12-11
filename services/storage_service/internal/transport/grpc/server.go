package grpc

import (
	"fmt"
	"log/slog"
	"net"

	storageGRPC "github.com/Nikita-Smirnov-idk/Antiplagiat_System/services/storage_service/internal/transport/grpc/storage"
	googleGRPC "google.golang.org/grpc"
)

type Server struct {
	log        *slog.Logger
	gRPCServer *googleGRPC.Server
	port       int
}

func New(log *slog.Logger, port int) *Server {
	gRPCServer := googleGRPC.NewServer()

	storageGRPC.Register(gRPCServer)

	return &Server{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *Server) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *Server) Run() error {
	const op = "grpcApp.Run"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	log.Info("starting gRPC server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Server) Stop() {
	const op = "grpcApp.Stop"

	a.log.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
