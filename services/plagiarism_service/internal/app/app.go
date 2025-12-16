package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/app/config"
	storageClient "github.com/Nikita-Smirnov-idk/plagiarism-service/internal/infrastructure/clients/storage"
	postgresRepo "github.com/Nikita-Smirnov-idk/plagiarism-service/internal/infrastructure/repositories/postgres"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/transport/grpc"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/use_cases"
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

	// Инициализация клиента storage-service
	storage, err := storageClient.NewStorageClient(cfg.Storage.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage client: %w", err)
	}
	log.Info("storage client initialized successfully")

	// Создание репозиториев
	dbRepository := postgresRepo.NewFileRepository(dbPool)

	// Создание use case сервиса
	plagiarismService := use_cases.NewPlagiarismService(log, dbRepository, storage.Client)

	// Инициализация gRPC сервера
	grpcApp := grpc.New(log, cfg.GRPC.Port, plagiarismService)

	return &App{
		GRPCSrv: grpcApp,
	}, nil
}

