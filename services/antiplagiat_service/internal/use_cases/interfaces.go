package use_cases

import (
	"context"

	"github.com/Nikita-Smirnov-idk/storage-service/internal/domain"
)

type S3Repository interface {
	GenerateUploadURL(key string) (string, error)
	VerifyUploadedFile(key string) error
	GenerateDownloadURL(key string, fromInside bool) (string, error)
}

type DBRepository interface {
	Save(ctx context.Context, file *domain.FileInfo) error
	GetByStudentAndTask(ctx context.Context, studentID, taskID string) (*domain.FileInfo, error)
	DeleteFile(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status domain.FileStatus) error
	ListTaskFiles(ctx context.Context, taskID string) ([]domain.FileInfo, error)
}
