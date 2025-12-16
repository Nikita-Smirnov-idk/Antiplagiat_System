package use_cases

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/domain"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/infrastructure/repositories"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/pkg/plagiarism_analyzer"
	storagepb "github.com/Nikita-Smirnov-idk/storage-service/contracts/gen/go"
	"github.com/google/uuid"
)

type PlagiarismService struct {
	logger  *slog.Logger
	db      DB
	storage storagepb.StorageClient
}

func NewPlagiarismService(logger *slog.Logger, db DB, storage storagepb.StorageClient) *PlagiarismService {
	return &PlagiarismService{
		logger:  logger,
		db:      db,
		storage: storage,
	}
}

func (s *PlagiarismService) GetPlagiarismReport(ctx context.Context, taskId string) (*Task, error) {
	const op = "Plagiarism_Service.GetPlagiarismReport"

	logger := s.logger.With(
		slog.String("op", op),
		slog.String("task_id", taskId),
	)

	task, err := s.db.GetTaskByID(ctx, taskId)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			logger.Info("task not found in DB, creating new task and running analysis")

			analysisTime := time.Now()
			newTask := &domain.Task{
				ID:                taskId,
				AnalysisStartedAt: analysisTime,
			}
			if err2 := s.db.SaveTask(ctx, newTask); err2 != nil {
				logger.Error("failed to save task", "error", err2)
				return nil, err2
			}

			response, err2 := s.storage.ListTaskFiles(
				ctx,
				&storagepb.ListTaskFilesRequest{
					TaskId: taskId,
				},
			)
			if err2 != nil {
				logger.Error("failed to contact storage", "error", err2)
				return nil, ErrExternalConnectionFailed
			}

			files := response.Items
			reports, err2 := s.runAnalysis(ctx, taskId, files, logger)
			if err2 != nil {
				return nil, err2
			}

			return &Task{
				ID:        taskId,
				StartedAt: analysisTime,
				Reports:   reports,
			}, nil
		} else {
			logger.Error("failed to load task", "error", err)
			return nil, err
		}
	}

	response, err := s.storage.ListTaskFiles(
		ctx,
		&storagepb.ListTaskFilesRequest{
			TaskId: taskId,
		},
	)

	if err != nil {
		logger.Error("failed to contact storage", "error", err)
		return nil, ErrExternalConnectionFailed
	}

	files := response.Items

	shouldReanalyze := false
	for _, f := range files {
		if f.UpdatedAt.AsTime().After(task.AnalysisStartedAt) {
			shouldReanalyze = true
			break
		}
	}

	if !shouldReanalyze {
		cachedReports, err2 := s.collectReportsFromCache(ctx, taskId, files, logger)
		if err2 != nil {
			return nil, err2
		}

		return &Task{
			ID:        task.ID,
			StartedAt: task.AnalysisStartedAt,
			Reports:   cachedReports,
		}, nil
	}

	reports, err := s.runAnalysis(ctx, taskId, files, logger)
	if err != nil {
		return nil, err
	}

	task.AnalysisStartedAt = time.Now()
	if err = s.db.UpdateTaskAnalysisTime(ctx, task.ID, task.AnalysisStartedAt); err != nil {
		logger.Error("failed to update task analysis time", "error", err)
		return nil, err
	}

	return &Task{
		ID:        task.ID,
		StartedAt: task.AnalysisStartedAt,
		Reports:   reports,
	}, nil
}

func toUseCaseReport(r domain.PlagiarismReport, currentStudent string) PlagiarismReport {
	report := PlagiarismReport{
		MaxSimilarity: r.Similarity,
	}

	// если текущий студент числится как student_a, то зеркально заполним иначе наоборот
	if r.StudentA == currentStudent {
		report.Student = r.StudentA
		report.StudentWithSimilarFile = r.StudentB
		report.FileHandedOverAt = r.FileAHandedOverAt
	} else {
		report.Student = r.StudentB
		report.StudentWithSimilarFile = r.StudentA
		report.FileHandedOverAt = r.FileBHandedOverAt
	}

	return report
}

// collectReportsFromCache собирает сохранённые ранее отчёты по текущей задаче.
func (s *PlagiarismService) collectReportsFromCache(
	ctx context.Context,
	taskID string,
	files []*storagepb.FileInfo,
	logger *slog.Logger,
) ([]PlagiarismReport, error) {
	return s.buildMaxReportsForTask(ctx, taskID, files, logger)
}

// runAnalysis сравнивает файлы попарно и сохраняет отчёты.
func (s *PlagiarismService) runAnalysis(
	ctx context.Context,
	taskID string,
	files []*storagepb.FileInfo,
	logger *slog.Logger,
) ([]PlagiarismReport, error) {
	if len(files) < 2 {
		return []PlagiarismReport{}, nil
	}

	// подготовим URL-ы для скачивания
	type fileData struct {
		studentID string
		updatedAt time.Time
		url       string
	}

	fileInfos := make([]fileData, 0, len(files))
	for _, f := range files {
		urlResp, err := s.storage.GenerateDownloadURL(ctx, &storagepb.GenerateDownloadURLRequest{
			StudentId:  f.GetStudentId(),
			TaskId:     taskID,
			FromInside: true,
		})
		if err != nil {
			logger.Error("failed to get download url", "student_id", f.GetStudentId(), "error", err)
			return nil, ErrExternalConnectionFailed
		}

		fileInfos = append(fileInfos, fileData{
			studentID: f.GetStudentId(),
			updatedAt: f.GetUpdatedAt().AsTime(),
			url:       urlResp.GetUrl(),
		})
	}

	if err := s.db.DeleteReportsByTaskID(ctx, taskID); err != nil {
		logger.Error("failed to delete old reports", "error", err)
		return nil, err
	}

	checker := plagiarism_analyzer.NewPlagiarismChecker(3, 0.7)

	for i := 0; i < len(fileInfos); i++ {
		for j := i + 1; j < len(fileInfos); j++ {
			fi := fileInfos[i]
			fj := fileInfos[j]

			similarity, err := checker.CompareFiles(fi.url, fj.url)
			if err != nil {
				logger.Error("failed to compare files", "student_a", fi.studentID, "student_b", fj.studentID, "error", err)
				return nil, &AnalysisError{
					StudentA: fi.studentID,
					StudentB: fj.studentID,
					Reason:   "file comparison failed",
					Err:      err,
				}
			}

			// сохраняем отчёт
			reportID, _ := uuid.NewUUID()
			dbReport := domain.PlagiarismReport{
				ID:                reportID,
				TaskId:            taskID,
				StudentA:          fi.studentID,
				StudentB:          fj.studentID,
				Similarity:        similarity,
				FileAHandedOverAt: fi.updatedAt,
				FileBHandedOverAt: fj.updatedAt,
			}

			if err = s.db.SaveReport(ctx, &dbReport); err != nil {
				logger.Error("failed to save report", "error", err)
				return nil, err
			}
		}
	}

	// после сохранения всех отчётов в БД возвращаем по одному (с максимальным совпадением) для каждого студента
	return s.buildMaxReportsForTask(ctx, taskID, files, logger)
}

// buildMaxReportsForTask выбирает для каждого студента отчёт с максимальным Similarity по указанной задаче.
func (s *PlagiarismService) buildMaxReportsForTask(
	ctx context.Context,
	taskID string,
	files []*storagepb.FileInfo,
	logger *slog.Logger,
) ([]PlagiarismReport, error) {
	result := make([]PlagiarismReport, 0)
	seen := make(map[string]bool)

	for _, f := range files {
		studentID := f.GetStudentId()
		if seen[studentID] {
			continue
		}
		seen[studentID] = true

		storedReports, err := s.db.GetReportsByStudentID(ctx, studentID)
		if err != nil {
			logger.Error("failed to load reports", "student_id", studentID, "error", err)
			return nil, err
		}

		var maxRep *domain.PlagiarismReport
		for i := range storedReports {
			r := &storedReports[i]
			if r.TaskId != taskID {
				continue
			}
			if maxRep == nil || r.Similarity > maxRep.Similarity {
				repCopy := *r
				maxRep = &repCopy
			}
		}

		if maxRep != nil {
			result = append(result, toUseCaseReport(*maxRep, studentID))
		}
	}

	return result, nil
}
