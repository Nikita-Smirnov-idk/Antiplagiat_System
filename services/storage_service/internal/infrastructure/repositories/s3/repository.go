package s3

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Repo struct {
	logger  *slog.Logger
	Storage Storage
}

func NewRepo(storage *Storage, logger *slog.Logger) *Repo {
	return &Repo{
		logger:  logger,
		Storage: *storage,
	}
}

func (s *Repo) GenerateUploadURL(key string) (string, error) {
	const op = "S3.REPO.GenerateUploadURL"

	logger := s.logger.With(
		slog.String("op", op),
		slog.String("file key", key),
	)

	req, _ := s.Storage.externalClient.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.Storage.bucket),
		Key:    aws.String(key),
	})

	conditions := []string{
		fmt.Sprintf(`["content-length-range", 1, %d]`, int64(5*math.Pow(10, 8))),
	}

	policy := fmt.Sprintf(`{
        "expiration": "%s",
        "conditions": [%s]
    }`,
		time.Now().Add(s.Storage.expirationTime).Format("2006-01-02T15:04:05Z"),
		strings.Join(conditions, ","),
	)

	encodedPolicy := base64.StdEncoding.EncodeToString([]byte(policy))

	if req.HTTPRequest.Form == nil {
		req.HTTPRequest.Form = make(url.Values)
	}

	req.HTTPRequest.Form.Add("Policy", encodedPolicy)

	presignedUrl, err := req.Presign(s.Storage.expirationTime)
	if err != nil {
		logger.Error("failed to generate presigned url", "error", err)
		return "", err
	}

	return presignedUrl, nil
}

func (s *Repo) VerifyUploadedFile(key string) error {
	const op = "S3.REPO.VerifyUploadedFile"

	logger := s.logger.With(
		slog.String("op", op),
		slog.String("file key", key),
	)

	_, err := s.Storage.internalClient.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.Storage.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error("failed to find file", "error", err)
		return fmt.Errorf("failed to find file: %w", err)
	}

	return nil
}

func (s *Repo) GenerateDownloadURL(key string, fromInside bool) (string, error) {
	const op = "S3.REPO.GenerateDownloadURL"

	logger := s.logger.With(
		slog.String("op", op),
		slog.String("file key", key),
	)

	var client *s3.S3

	if fromInside {
		client = s.Storage.internalClient
	} else {
		client = s.Storage.externalClient
	}

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.Storage.bucket),
		Key:    aws.String(key),
	})

	presignedUrl, err := req.Presign(s.Storage.expirationTime)
	if err != nil {
		logger.Error("failed to generate presigned url", "error", err)
		return "", err
	}

	return presignedUrl, nil
}
