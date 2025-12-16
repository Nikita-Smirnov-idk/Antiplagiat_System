package app

import (
	"context"
	"log/slog"

	"api_gateway/internal/app/config"
	"api_gateway/internal/transport/http"

	plagiarismpb "github.com/Nikita-Smirnov-idk/plagiarism-service/contracts/gen/go"
	storagepb "github.com/Nikita-Smirnov-idk/storage-service/contracts/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	HTTPServer *http.Server
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) (*App, error) {
	storageConn, err := grpc.DialContext(ctx, cfg.Storage.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	storageClient := storagepb.NewStorageClient(storageConn)

	analysisConn, err := grpc.DialContext(ctx, cfg.Analysis.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	analysisClient := plagiarismpb.NewPlagiarismClient(analysisConn)

	httpServer := http.NewServer(log, cfg.HTTP.Port, storageClient, analysisClient)

	return &App{
		HTTPServer: httpServer,
	}, nil
}


