package use_cases

import "errors"

var (
	ErrFailedToGenerateURL = errors.New("failed to generate url")
	ErrFileNotFound        = errors.New("file not found")
	ErrFileYetNotUploaded  = errors.New("file has not been uploaded yet")
)
