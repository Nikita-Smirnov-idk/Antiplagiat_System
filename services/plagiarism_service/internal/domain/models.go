package domain

import (
	"time"

	"github.com/google/uuid"
)

type PlagiarismReport struct {
	ID                uuid.UUID `json:"id" db:"id"`
	TaskId            string    `json:"task_id" db:"task_id"`
	StudentA          string    `json:"student_a" db:"student_a"`
	StudentB          string    `json:"student_b" db:"student_b"`
	Similarity        float64   `json:"similarity" db:"similarity"`
	FileAHandedOverAt time.Time `json:"file_a_handed_over_at" db:"file_a_handed_over_at"`
	FileBHandedOverAt time.Time `json:"file_b_handed_over_at" db:"file_b_handed_over_at"`
}

type Task struct {
	ID                string
	AnalysisStartedAt time.Time
	Reports           []PlagiarismReport
}
