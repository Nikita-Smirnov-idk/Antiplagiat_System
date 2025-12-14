package use_cases

import "errors"

var (
	ErrFailedToGenerateURL = errors.New("failed to generate url")
	ErrFileNotFound        = errors.New("file not found")
	ErrorTaskNotFound      = errors.New("task not found")
)
