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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service interface {
	GenerateUploadURL(ctx context.Context, studentId, taskId string) (string, error)
	VerifyUploadedFile(ctx context.Context, studentId, taskId string) (string, error)
	GenerateDownloadURL(ctx context.Context, studentId, taskId string, fromInside bool) (string, error)
	ListTaskFiles(ctx context.Context, taskID string) ([]use_cases.SafeFileInfo, error)
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

	defer func() {
		if r := recover(); r != nil {
			logger.Error(
				"PANIC",
				"recover", r,
			)
		}
	}()

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
			return nil, status.Error(codes.Internal, "failed to generate url")
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

	defer func() {
		if r := recover(); r != nil {
			logger.Error(
				"PANIC",
				"recover", r,
			)
		}
	}()

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
		slog.Bool("FromInside", req.GetFromInside()),
	)

	defer func() {
		if r := recover(); r != nil {
			logger.Error(
				"PANIC",
				"recover", r,
			)
		}
	}()

	if req.GetStudentId() == "" || req.GetTaskId() == "" {
		logger.Warn("StudentId and TaskId were not given")
		return nil, status.Error(
			codes.InvalidArgument,
			"student_id and task_id are required",
		)
	}
	url, err := h.service.GenerateDownloadURL(ctx, req.GetStudentId(), req.GetTaskId(), req.GetFromInside())

	if err != nil {
		if errors.Is(err, use_cases.ErrFileNotFound) {
			logger.Error("file not found", "error", err)
			return nil, status.Error(codes.NotFound, "file not found")
		}
		if errors.Is(err, use_cases.ErrFailedToGenerateURL) {
			logger.Error("failed to generate url", "error", err)
			return nil, status.Error(codes.Internal, "failed to generate url")
		}
		if errors.Is(err, use_cases.ErrFileYetNotUploaded) {
			logger.Error("file has not been uploaded yet", "error", err)
			return nil, status.Error(codes.FailedPrecondition, "file has not been uploaded yet")
		}

		logger.Error("internal error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &gen.GenerateDownloadURLResponse{
		Url: url,
	}, nil
}

func (h *Handler) ListTaskFiles(ctx context.Context, req *gen.ListTaskFilesRequest) (*gen.ListTaskFilesResponse, error) {
	const op = "Handler.ListTaskFiles"

	logger := h.logger.With(
		slog.String("op", op),
		slog.String("TaskId", req.GetTaskId()),
	)

	defer func() {
		if r := recover(); r != nil {
			logger.Error(
				"PANIC",
				"recover", r,
			)
		}
	}()

	if req.GetTaskId() == "" {
		logger.Warn("TaskId was not given")
		return nil, status.Error(
			codes.InvalidArgument,
			"task_id is required",
		)
	}
	files, err := h.service.ListTaskFiles(ctx, req.GetTaskId())

	if err != nil {
		logger.Error("internal error", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	var result []*gen.FileInfo

	for _, file := range files {
		var updatedAt *timestamppb.Timestamp
		if !file.UpdatedAt.IsZero() {
			updatedAt = timestamppb.New(file.UpdatedAt)
		}

		result = append(result, &gen.FileInfo{
			StudentId: file.StudentId,
			UpdatedAt: updatedAt,
			Status:    file.Status,
		})
	}

	return &gen.ListTaskFilesResponse{
		Items: result,
	}, nil
}
