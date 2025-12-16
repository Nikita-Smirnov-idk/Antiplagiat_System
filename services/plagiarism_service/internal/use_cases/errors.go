package use_cases

import (
	"errors"
	"fmt"
)

var (
	ErrExternalConnectionFailed = errors.New("failed to connect external service")
	ErrFileExtractionFailed     = errors.New("failed to extract text from file")
	ErrFileDownloadFailed       = errors.New("failed to download file")
)

type AnalysisError struct {
	StudentA string
	StudentB string
	Reason   string
	Err      error
}

func (e *AnalysisError) Error() string {
	if e.StudentA != "" && e.StudentB != "" {
		return fmt.Sprintf("analysis failed for students %s and %s: %s: %v", e.StudentA, e.StudentB, e.Reason, e.Err)
	}
	return fmt.Sprintf("analysis failed: %s: %v", e.Reason, e.Err)
}

func (e *AnalysisError) Unwrap() error {
	return e.Err
}
