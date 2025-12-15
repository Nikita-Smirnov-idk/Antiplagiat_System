package use_cases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Nikita-Smirnov-idk/storage-service/internal/domain"
	"github.com/Nikita-Smirnov-idk/storage-service/internal/infrastructure/repositories"
	"github.com/google/uuid"
)

type FileService struct {
	logger *slog.Logger
	S3     S3Repository
	DB     DBRepository
}

func NewFileService(s3 S3Repository, db DBRepository, logger *slog.Logger) *FileService {
	return &FileService{
		S3:     s3,
		DB:     db,
		logger: logger,
	}
}

func (f *FileService) GenerateUploadURL(ctx context.Context, studentId, taskId string) (string, error) {
	const op = "Storage_Service.GenerateUploadURL"

	logger := f.logger.With(
		slog.String("op", op),
	)

	fileInfo, err := f.DB.GetByStudentAndTask(ctx, studentId, taskId)

	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			fileId, uuidErr := uuid.NewUUID()
			status := domain.FileStatusUploading
			updatedAt := time.Now()

			if uuidErr != nil {
				logger.Error("failed to generate uuid", "error", uuidErr)
				return "", fmt.Errorf("failed to generate uuid: %w", uuidErr)
			}

			fileInfo = domain.NewFileInfo(fileId, studentId, taskId, updatedAt, status)

			saveErr := f.DB.Save(ctx, fileInfo)
			if saveErr != nil {
				logger.Error("failed to save data to database", "error", saveErr)
				return "", fmt.Errorf("failed to save data to database: %w", saveErr)
			}
		} else {
			logger.Error("failed to find file", "error", err)
			return "", fmt.Errorf("failed to find file: %w", err)
		}
	}

	logger.Info("File Info", "file id", fileInfo.ID.String())

	urlToUpload, err := f.S3.GenerateUploadURL(fileInfo.TaskID + "/" + fileInfo.ID.String())

	if err != nil {
		logger.Error("failed to generate url", "error", err)
		return "", ErrFailedToGenerateURL
	}

	return urlToUpload, nil
}

func (f *FileService) VerifyUploadedFile(ctx context.Context, studentId, taskId string) (string, error) {
	const op = "Storage_Service.VerifyUploadedFile"

	logger := f.logger.With(
		slog.String("op", op),
	)

	fileInfo, err := f.DB.GetByStudentAndTask(ctx, studentId, taskId)
	if err != nil {
		logger.Error("failed to find file", "error", err)
		return "", ErrFileNotFound
	}

	logger.Info("File Info", "file id", fileInfo.ID.String())

	err = f.S3.VerifyUploadedFile(fileInfo.TaskID + "/" + fileInfo.ID.String())

	if err != nil {
		logger.Error("failed to verify uploaded file", "error", err)
		return "", fmt.Errorf("failed to verify uploaded file: %w", err)
	}

	err = f.DB.UpdateStatus(ctx, fileInfo.ID.String(), domain.FileStatusUploaded)
	if err != nil {
		logger.Error("failed to update status", "error", err)
		return "", fmt.Errorf("failed to update status: %w", err)
	}

	return fileInfo.ID.String(), nil
}

func (f *FileService) GenerateDownloadURL(ctx context.Context, studentId, taskId string, fromInside bool) (string, error) {
	const op = "Storage_Service.GenerateDownloadURL"

	logger := f.logger.With(
		slog.String("op", op),
	)

	fileInfo, err := f.DB.GetByStudentAndTask(ctx, studentId, taskId)
	if err != nil {
		logger.Error("failed to find file", "error", err)
		return "", ErrFileNotFound
	}

	if fileInfo.Status != domain.FileStatusUploaded {
		logger.Error("fail has not been uploaded yet", "error", err)
		return "", ErrFileYetNotUploaded
	}

	logger.Info("File Info", "file id", fileInfo.ID.String())

	urlToDownload, err := f.S3.GenerateDownloadURL(fileInfo.TaskID+"/"+fileInfo.ID.String(), fromInside)

	if err != nil {
		logger.Error("failed to generate url", "error", err)
		return "", ErrFailedToGenerateURL
	}

	return urlToDownload, nil
}

func (f *FileService) ListTaskFiles(ctx context.Context, taskID string) ([]SafeFileInfo, error) {
	const op = "Storage_Service.ListTaskFiles"

	logger := f.logger.With(
		slog.String("op", op),
	)

	files, err := f.DB.ListTaskFiles(ctx, taskID)
	if err != nil {
		logger.Error("failed to find files by task id", "error", err)
		return nil, fmt.Errorf("failed to find files by task id: %w", err)
	}

	var result []SafeFileInfo

	for _, file := range files {
		var item SafeFileInfo

		item = SafeFileInfo{
			StudentId: file.StudentID,
			UpdatedAt: file.UpdatedAt,
			Status:    string(file.Status),
		}

		result = append(result, item)
	}

	return result, nil
}
