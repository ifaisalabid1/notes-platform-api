package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

type R2Storage struct {
	client     *s3.Client
	bucketName string
}

func NewR2Storage(ctx context.Context, cfg R2Config) (*R2Storage, error) {
	if strings.TrimSpace(cfg.AccountID) == "" {
		return nil, errors.New("r2 account id is required")
	}

	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return nil, errors.New("r2 access key id is required")
	}

	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, errors.New("r2 secret access key is required")
	}

	if strings.TrimSpace(cfg.BucketName) == "" {
		return nil, errors.New("r2 bucket name is required")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config for r2: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})

	return &R2Storage{
		client:     client,
		bucketName: cfg.BucketName,
	}, nil
}

func (s *R2Storage) PutObject(ctx context.Context, input PutObjectInput) (PutObjectResult, error) {
	if strings.TrimSpace(input.Key) == "" {
		return PutObjectResult{}, errors.New("storage key is required")
	}

	if input.Body == nil {
		return PutObjectResult{}, errors.New("storage body is required")
	}

	cleanKey, err := safeObjectKey(input.Key)
	if err != nil {
		return PutObjectResult{}, err
	}

	bodyBytes, err := io.ReadAll(input.Body)
	if err != nil {
		return PutObjectResult{}, fmt.Errorf("read object body: %w", err)
	}

	if len(bodyBytes) == 0 {
		return PutObjectResult{}, errors.New("object body is empty")
	}

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(cleanKey),
		Body:        bytes.NewReader(bodyBytes),
		ContentType: aws.String(input.ContentType),
	})
	if err != nil {
		return PutObjectResult{}, fmt.Errorf("put r2 object: %w", err)
	}

	return PutObjectResult{
		Key:       cleanKey,
		SizeBytes: int64(len(bodyBytes)),
	}, nil
}

func (s *R2Storage) DeleteObject(ctx context.Context, key string) error {
	cleanKey, err := safeObjectKey(key)
	if err != nil {
		return err
	}

	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(cleanKey),
	})
	if err != nil {
		return fmt.Errorf("delete r2 object: %w", err)
	}

	return nil
}
