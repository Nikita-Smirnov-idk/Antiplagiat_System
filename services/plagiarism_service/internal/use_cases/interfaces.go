package use_cases

import (
	"context"
	"time"

	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/domain"
)

type DB interface {
	GetReportsByStudentID(ctx context.Context, studentID string) ([]domain.PlagiarismReport, error)
	GetTaskByID(ctx context.Context, taskID string) (*domain.Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	DeleteReportsByTaskID(ctx context.Context, taskID string) error
	SaveReport(ctx context.Context, report *domain.PlagiarismReport) error
	SaveTask(ctx context.Context, task *domain.Task) error
	UpdateTaskAnalysisTime(ctx context.Context, taskID string, analysisStartedAt time.Time) error
}
