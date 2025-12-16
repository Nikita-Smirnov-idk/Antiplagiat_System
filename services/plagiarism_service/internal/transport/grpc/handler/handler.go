package handler

import (
	"context"
	"log/slog"

	gen "github.com/Nikita-Smirnov-idk/plagiarism-service/contracts/gen/go"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/use_cases"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PlagiarismService interface {
	GetPlagiarismReport(ctx context.Context, taskId string) (use_cases.Task, error)
}

type Handler struct {
	gen.UnimplementedPlagiarismServer
	service PlagiarismService
	logger  *slog.Logger
}

func Register(gRPC *grpc.Server, service PlagiarismService, logger *slog.Logger) {
	gen.RegisterPlagiarismServer(gRPC, &Handler{service: service, logger: logger})
}

func (h *Handler) GetPlagiarismReport(ctx context.Context, req *gen.GetPlagiarismReportRequest) (*gen.GetPlagiarismReportResponse, error) {
	const op = "Handler.GenerateUploadURL"

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

	err := ValidateTaskId(req.GetTaskId(), h.logger)
	if err != nil {
		return nil, err
	}

	taskReport, err := h.service.GetPlagiarismReport(ctx, req.GetTaskId())

	if err != nil {
		return nil, err
	}

	var result []*gen.PlagiarismReport

	for _, report := range taskReport.Reports {
		var fileHandedOverAt *timestamppb.Timestamp

		if !report.FileHandedOverAt.IsZero() {
			fileHandedOverAt = timestamppb.New(report.FileHandedOverAt)
		}

		result = append(result, &gen.PlagiarismReport{
			Student:                report.Student,
			StudentWithSimilarFile: report.StudentWithSimilarFile,
			MaxSimilarity:          report.MaxSimilarity,
			FileHandedOverAt:       fileHandedOverAt,
		})
	}

	var startedAt *timestamppb.Timestamp

	if !taskReport.StartedAt.IsZero() {
		startedAt = timestamppb.New(taskReport.StartedAt)
	}

	return &gen.GetPlagiarismReportResponse{
		Reports:   result,
		StartedAt: startedAt,
	}, nil

}
