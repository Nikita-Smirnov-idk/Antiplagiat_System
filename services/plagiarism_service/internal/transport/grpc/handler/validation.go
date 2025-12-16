package handler

import (
	"errors"
	"log/slog"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateTaskId(taskId string, log *slog.Logger) error {
	const op = "Handler.Validation.ValidateTaskId"

	logger := log.With(
		slog.String("op", op),
		slog.String("id", taskId),
	)

	err := ValidateIdWrapped(taskId, "task", logger)
	if err != nil {
		return err
	}

	return nil
}

func ValidateIdWrapped(id, nameOfId string, logger *slog.Logger) error {
	err := ValidateId(id)
	if err != nil {
		if errors.Is(err, ErrIdRequired) {
			logger.Warn(nameOfId + " id required")
			return status.Error(
				codes.InvalidArgument,
				nameOfId+" id required",
			)
		}
		if errors.Is(err, ErrIdTooLong) {
			logger.Warn(nameOfId + " id is too long")
			return status.Error(
				codes.InvalidArgument,
				nameOfId+" id is too long",
			)
		}
		logger.Warn("invalid argument")
		return status.Error(
			codes.InvalidArgument,
			"invalid argument",
		)
	}

	return nil
}

func ValidateId(id string) error {
	if id == "" {
		return ErrIdRequired
	}

	if utf8.RuneCountInString(id) > 50 {
		return ErrIdTooLong
	}
	return nil
}
