package s3

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
)

type Storage struct {
	bucket         string
	region         string
	endpoint       string
	expirationTime time.Duration
	internalClient *s3.S3
	externalClient *s3.S3
}

func NewStorage(region, bucket, endpoint, clientEndpoint, secret, access string, expirationTime int32) (*Storage, error) {
	internalClient, err := NewClient(region, endpoint, secret, access)

	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client")
	}

	externalClient, err := NewClient(region, clientEndpoint, secret, access)

	if err != nil {
		return nil, fmt.Errorf("failed to create s3 client")
	}

	return &Storage{
		bucket:         bucket,
		region:         region,
		endpoint:       endpoint,
		expirationTime: time.Duration(expirationTime) * time.Minute,
		internalClient: internalClient.client,
		externalClient: externalClient.client,
	}, nil
}
