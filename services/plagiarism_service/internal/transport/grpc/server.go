package grpc

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/transport/grpc/handler"
	googleGRPC "google.golang.org/grpc"
)

type Server struct {
	logger     *slog.Logger
	gRPCServer *googleGRPC.Server
	port       int
}

func New(logger *slog.Logger, port int, service handler.Service) *Server {
	gRPCServer := googleGRPC.NewServer()

	handler.Register(gRPCServer, service, logger)

	return &Server{
		logger:     logger,
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
	const op = "server.Run"

	logger := a.logger.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	logger.Info("starting gRPC server")

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	logger.Info("grpc server is running", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *Server) Stop() {
	const op = "server.Stop"

	a.logger.With(slog.String("op", op)).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
