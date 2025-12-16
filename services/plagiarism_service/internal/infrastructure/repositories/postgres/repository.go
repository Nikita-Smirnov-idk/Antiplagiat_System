package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/domain"
	"github.com/Nikita-Smirnov-idk/plagiarism-service/internal/infrastructure/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{pool: pool}
}

func (r *FileRepo) SaveTask(ctx context.Context, task *domain.Task) error {
	query := `INSERT INTO tasks (id, analysis_started_at) VALUES ($1, $2) 
	          ON CONFLICT (id) DO UPDATE SET analysis_started_at = EXCLUDED.analysis_started_at`

	_, err := r.pool.Exec(ctx, query, task.ID, task.AnalysisStartedAt)
	return err
}

func (r *FileRepo) SaveReport(ctx context.Context, report *domain.PlagiarismReport) error {
	query := `INSERT INTO plagiarism_reports 
	          (id, task_id, student_a, student_b, similarity, file_a_handed_over_at, file_b_handed_over_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		report.ID.String(),
		report.TaskId,
		report.StudentA,
		report.StudentB,
		report.Similarity,
		report.FileAHandedOverAt,
		report.FileBHandedOverAt)
	return err
}

func (r *FileRepo) DeleteTask(ctx context.Context, taskID string) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, taskID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return nil
	}

	return nil
}

func (r *FileRepo) DeleteReportsByTaskID(ctx context.Context, taskID string) error {
	query := `DELETE FROM plagiarism_reports WHERE task_id = $1`

	_, err := r.pool.Exec(ctx, query, taskID)
	return err
}

func (r *FileRepo) GetReportsByStudentID(ctx context.Context, studentID string) ([]domain.PlagiarismReport, error) {
	query := `SELECT id, task_id, student_a, student_b, similarity, file_a_handed_over_at, file_b_handed_over_at 
	          FROM plagiarism_reports 
	          WHERE student_a = $1 OR student_b = $1`

	rows, err := r.pool.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []domain.PlagiarismReport
	for rows.Next() {
		var report domain.PlagiarismReport
		err := rows.Scan(
			&report.ID,
			&report.TaskId,
			&report.StudentA,
			&report.StudentB,
			&report.Similarity,
			&report.FileAHandedOverAt,
			&report.FileBHandedOverAt,
		)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *FileRepo) GetTaskByID(ctx context.Context, taskID string) (*domain.Task, error) {
	query := `SELECT id, analysis_started_at FROM tasks WHERE id = $1`

	var task domain.Task
	err := r.pool.QueryRow(ctx, query, taskID).Scan(&task.ID, &task.AnalysisStartedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	return &task, nil
}

// UpdateTaskAnalysisTime обновляет только время начала анализа для задачи.
func (r *FileRepo) UpdateTaskAnalysisTime(ctx context.Context, taskID string, analysisStartedAt time.Time) error {
	query := `UPDATE tasks SET analysis_started_at = $2 WHERE id = $1`

	_, err := r.pool.Exec(ctx, query, taskID, analysisStartedAt)
	return err
}
