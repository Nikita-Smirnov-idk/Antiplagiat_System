package domain

import (
	"time"

	"github.com/google/uuid"
)

type FileInfo struct {
	ID uuid.UUID `json:"id" db:"id"`

	StudentID string `json:"student_id" db:"student_id"`
	TaskID    string `json:"task_id" db:"task_id"`

	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Status    FileStatus `json:"status" db:"status"`
}

type FileStatus string

const (
	FileStatusUploading FileStatus = "uploading"
	FileStatusUploaded  FileStatus = "uploaded"
)

func NewFileInfo(id uuid.UUID, studentId, taskId string, updatedAt time.Time, status FileStatus) *FileInfo {
	return &FileInfo{
		ID:        id,
		StudentID: studentId,
		TaskID:    taskId,
		UpdatedAt: updatedAt,
		Status:    status,
	}
}
