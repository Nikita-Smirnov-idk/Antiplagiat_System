package use_cases

import (
	"context"
	"log/slog"
)

type PlagiarismService struct {
	logger *slog.Logger
}

func (s *PlagiarismService) GetPlagiarismReport(ctx context.Context, taskId string) (Task, error) {

}
