package handler

import (
	"context"
	"errors"
	"log/slog"

	gen "github.com/Nikita-Smirnov-idk/storage-service/contracts/gen/go"
	"github.com/Nikita-Smirnov-idk/storage-service/internal/use_cases"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	GenerateUploadURL(ctx context.Context, studentId, taskId string) (string, error)
	VerifyUploadedFile(ctx context.Context, studentId, taskId string) (string, error)
	GenerateDownloadURL(ctx context.Context, studentId, taskId string) (string, error)
	GetStudentIDsByTaskID(ctx context.Context, taskID string) ([]string, error)
}

type Handler struct {
	gen.UnimplementedStorageServer
	service Service
	logger  *slog.Logger
}

func Register(gRPC *grpc.Server, service Service, logger *slog.Logger) {
	gen.RegisterStorageServer(gRPC, &Handler{service: service, logger: logger})
}

func (h *Handler) GenerateUploadURL(ctx context.Context, req *gen.GenerateUploadURLRequest) (*gen.GenerateUploadURLResponse, error) {
	const op = "Handler.GenerateUploadURL"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("StudentId", req.GetStudentId()),
		slog.String("TaskId", req.GetTaskId()),
	)

	if req.GetStudentId() == "" || req.GetTaskId() == "" {
		logger.Warn("StudentId and TaskId were not given")
		return nil, status.Error(
			codes.InvalidArgument,
			"StudentId and TaskId are required",
		)
	}
	url, err := h.service.GenerateUploadURL(ctx, req.GetStudentId(), req.GetTaskId())

	if err != nil {
		if errors.Is(err, use_cases.ErrFailedToGenerateURL) {
			logger.Error("failed to generate url", "error", err)
			return nil, status.Error(codes.Unavailable, "failed to generate url")
		}

		logger.Error("internal error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &gen.GenerateUploadURLResponse{
		Url: url,
	}, nil
}
func (h *Handler) VerifyUploadedFile(ctx context.Context, req *gen.VerifyUploadedFileRequest) (*gen.VerifyUploadedFileResponse, error) {
	const op = "Handler.VerifyUploadedFile"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("StudentId", req.GetStudentId()),
		slog.String("TaskId", req.GetTaskId()),
	)

	if req.GetStudentId() == "" || req.GetTaskId() == "" {
		logger.Warn("StudentId and TaskId were not given")
		return nil, status.Error(
			codes.InvalidArgument,
			"student_id and task_id are required",
		)
	}
	fileId, err := h.service.VerifyUploadedFile(ctx, req.GetStudentId(), req.GetTaskId())

	if err != nil {
		if errors.Is(err, use_cases.ErrFileNotFound) {
			logger.Error("file not found", "error", err)
			return nil, status.Error(codes.NotFound, "file not found")
		}

		logger.Error("internal error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &gen.VerifyUploadedFileResponse{
		FileId: fileId,
	}, nil
}
func (h *Handler) GenerateDownloadURL(ctx context.Context, req *gen.GenerateDownloadURLRequest) (*gen.GenerateDownloadURLResponse, error) {
	const op = "Handler.GenerateDownloadURL"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("StudentId", req.GetStudentId()),
		slog.String("TaskId", req.GetTaskId()),
	)

	if req.GetStudentId() == "" || req.GetTaskId() == "" {
		logger.Warn("StudentId and TaskId were not given")
		return nil, status.Error(
			codes.InvalidArgument,
			"student_id and task_id are required",
		)
	}
	url, err := h.service.GenerateDownloadURL(ctx, req.GetStudentId(), req.GetTaskId())

	if err != nil {
		if errors.Is(err, use_cases.ErrFileNotFound) {
			logger.Error("file not found", "error", err)
			return nil, status.Error(codes.NotFound, "file not found")
		}
		if errors.Is(err, use_cases.ErrFailedToGenerateURL) {
			logger.Error("failed to generate url", "error", err)
			return nil, status.Error(codes.Unavailable, "failed to generate url")
		}

		logger.Error("internal error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &gen.GenerateDownloadURLResponse{
		Url: url,
	}, nil
}

func (h *Handler) GetStudentsByTaskId(ctx context.Context, req *gen.GetStudentsByTaskIdRequest) (*gen.GetStudentsByTaskIdResponse, error) {
	const op = "Handler.GetStudentIDsByTaskID"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("TaskId", req.GetTaskId()),
	)

	if req.GetTaskId() == "" {
		logger.Warn("TaskId was not given")
		return nil, status.Error(
			codes.InvalidArgument,
			"task_id is required",
		)
	}
	studentsIds, err := h.service.GetStudentIDsByTaskID(ctx, req.GetTaskId())

	if err != nil {
		if errors.Is(err, use_cases.ErrFileNotFound) {
			logger.Error("file not found", "error", err)
			return nil, status.Error(codes.NotFound, "file not found")
		}
		if errors.Is(err, use_cases.ErrFailedToGenerateURL) {
			logger.Error("failed to generate url", "error", err)
			return nil, status.Error(codes.Unavailable, "failed to generate url")
		}

		logger.Error("internal error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &gen.GetStudentsByTaskIdResponse{
		StudentIds: studentsIds,
	}, nil
}
