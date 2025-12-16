package use_cases

import (
	"time"
)

type PlagiarismReport struct {
	Student                string
	StudentWithSimilarFile string
	MaxSimilarity          float64
	FileHandedOverAt       time.Time
}

type Task struct {
	ID        string
	StartedAt time.Time
	Reports   []PlagiarismReport
}
