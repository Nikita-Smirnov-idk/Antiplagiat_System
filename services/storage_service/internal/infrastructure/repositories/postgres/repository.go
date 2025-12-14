package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Nikita-Smirnov-idk/storage-service/internal/domain"
	"github.com/Nikita-Smirnov-idk/storage-service/internal/infrastructure/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepo struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepo {
	return &FileRepo{pool: pool}
}

func (r *FileRepo) Save(ctx context.Context, file *domain.FileInfo) error {
	query := `
        INSERT INTO files (id, student_id, task_id, updated_at, status)
        VALUES ($1, $2, $3, $4, $5)
    `

	_, err := r.pool.Exec(ctx, query,
		file.ID,
		file.StudentID,
		file.TaskID,
		file.UpdatedAt,
		string(file.Status),
	)

	return err
}

func (r *FileRepo) GetStudentIDsByTaskID(ctx context.Context, taskID string) ([]string, error) {
	query := `
		SELECT DISTINCT student_id
		FROM files 
		WHERE task_id = $1
		ORDER BY student_id
	`

	rows, err := r.pool.Query(ctx, query, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get student IDs: %w", err)
	}
	defer rows.Close()

	var studentIDs []string
	for rows.Next() {
		var studentID string
		if err := rows.Scan(&studentID); err != nil {
			return nil, fmt.Errorf("failed to scan student ID: %w", err)
		}
		studentIDs = append(studentIDs, studentID)
	}

	return studentIDs, nil
}

func (r *FileRepo) GetByID(ctx context.Context, id string) (*domain.FileInfo, error) {
	query := `
		SELECT *
		FROM files 
		WHERE id = $1
	`

	return r.scanFile(ctx, query, id)
}

func (r *FileRepo) GetByStudentAndTask(ctx context.Context, studentID, taskID string) (*domain.FileInfo, error) {
	query := `
		SELECT *
		FROM files 
		WHERE student_id = $1 AND task_id = $2
	`

	return r.scanFile(ctx, query, studentID, taskID)
}

func (r *FileRepo) GetByTask(ctx context.Context, taskID string) (*domain.FileInfo, error) {
	query := `
		SELECT *
		FROM files 
		WHERE task_id = $2
	`

	return r.scanFile(ctx, query, taskID)
}

func (r *FileRepo) DeleteFile(ctx context.Context, id string) error {
	query := `
		DELETE FROM files 
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("file with id %s not found", id)
	}

	return nil
}

func (r *FileRepo) UpdateStatus(ctx context.Context, id string, status domain.FileStatus) error {
	query := `
		UPDATE files 
		SET status = $2, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, string(status))
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("file with id %s not found", id)
	}

	return nil
}

func (r *FileRepo) scanFile(ctx context.Context, query string, args ...interface{}) (*domain.FileInfo, error) {
	var file domain.FileInfo
	var status string

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&file.ID,
		&file.StudentID,
		&file.TaskID,
		&file.UpdatedAt,
		&status,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repositories.ErrNotFound
		}
		return nil, err
	}

	file.Status = domain.FileStatus(status)

	return &file, nil
}
