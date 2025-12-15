package use_cases

import (
	"time"
)

type SafeFileInfo struct {
	StudentId string `json:"student_id"`

	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status"`
}
