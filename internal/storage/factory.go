package storage

import (
	"context"
	"fmt"
)

type Driver string

const (
	DriverLocal Driver = "local"
	DriverR2    Driver = "r2"
)

type FactoryConfig struct {
	Driver string

	LocalStorageDir string

	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
}

func NewFromConfig(ctx context.Context, cfg FactoryConfig) (ObjectStorage, error) {
	switch Driver(cfg.Driver) {
	case DriverLocal:
		return NewLocalStorage(cfg.LocalStorageDir), nil

	case DriverR2:
		return NewR2Storage(ctx, R2Config{
			AccountID:       cfg.R2AccountID,
			AccessKeyID:     cfg.R2AccessKeyID,
			SecretAccessKey: cfg.R2SecretAccessKey,
			BucketName:      cfg.R2BucketName,
		})

	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Driver)
	}
}
