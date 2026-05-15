package storage

import (
	"context"
	"io"
)

type PutObjectInput struct {
	Key         string
	Body        io.Reader
	ContentType string
}

type PutObjectResult struct {
	Key       string
	SizeBytes int64
}

type ObjectStorage interface {
	PutObject(ctx context.Context, input PutObjectInput) (PutObjectResult, error)
	DeleteObject(ctx context.Context, key string) error
}
