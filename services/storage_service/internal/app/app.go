package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nikita-Smirnov-idk/storage-service/internal/app/config"
	postgresRepo "github.com/Nikita-Smirnov-idk/storage-service/internal/infrastructure/repositories/postgres"
	s3Repo "github.com/Nikita-Smirnov-idk/storage-service/internal/infrastructure/repositories/s3"
	"github.com/Nikita-Smirnov-idk/storage-service/internal/transport/grpc"
	"github.com/Nikita-Smirnov-idk/storage-service/internal/use_cases"
)

type App struct {
	GRPCSrv *grpc.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	// Инициализация PostgreSQL
	dbPool, err := postgresRepo.New(
		ctx,
		cfg.DB.AutoMigrate,
		cfg.DB.Path,
		cfg.DB.MinConn,
		cfg.DB.MaxConn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres: %w", err)
	}
	log.Info("postgres initialized successfully")

	// Инициализация S3
	s3Storage, err := s3Repo.NewStorage(
		cfg.S3.Region,
		cfg.S3.Bucket,
		cfg.S3.Endpoint,
		cfg.S3.ClientEndpoint,
		cfg.S3.SecretKey,
		cfg.S3.AccessKey,
		cfg.S3.ExpirationTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3: %w", err)
	}
	log.Info("s3 storage initialized successfully")

	// Создание репозиториев
	dbRepository := postgresRepo.NewFileRepository(dbPool)
	s3Repository := s3Repo.NewRepo(s3Storage, log)

	// Создание use case сервиса
	fileService := use_cases.NewFileService(s3Repository, dbRepository, log)

	// Инициализация gRPC сервера
	grpcApp := grpc.New(log, cfg.GRPC.Port, fileService)

	return &App{
		GRPCSrv: grpcApp,
	}, nil
}
