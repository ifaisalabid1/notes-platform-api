package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	baseDir string
}

func NewLocalStorage(baseDir string) *LocalStorage {
	return &LocalStorage{
		baseDir: baseDir,
	}
}

func (s *LocalStorage) PutObject(ctx context.Context, input PutObjectInput) (PutObjectResult, error) {
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

	finalPath := filepath.Join(s.baseDir, cleanKey)

	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return PutObjectResult{}, fmt.Errorf("create storage directory: %w", err)
	}

	tempFile, err := os.CreateTemp(filepath.Dir(finalPath), ".upload-*")
	if err != nil {
		return PutObjectResult{}, fmt.Errorf("create temp object: %w", err)
	}

	tempPath := tempFile.Name()

	var written int64

	copyDone := make(chan error, 1)

	go func() {
		defer tempFile.Close()

		n, copyErr := io.Copy(tempFile, input.Body)
		written = n

		if copyErr != nil {
			copyDone <- copyErr
			return
		}

		if syncErr := tempFile.Sync(); syncErr != nil {
			copyDone <- syncErr
			return
		}

		copyDone <- nil
	}()

	select {
	case <-ctx.Done():
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return PutObjectResult{}, ctx.Err()
	case err := <-copyDone:
		if err != nil {
			_ = os.Remove(tempPath)
			return PutObjectResult{}, fmt.Errorf("write object: %w", err)
		}
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		_ = os.Remove(tempPath)
		return PutObjectResult{}, fmt.Errorf("finalize object: %w", err)
	}

	return PutObjectResult{
		Key:       cleanKey,
		SizeBytes: written,
	}, nil
}

func (s *LocalStorage) DeleteObject(ctx context.Context, key string) error {
	cleanKey, err := safeObjectKey(key)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	path := filepath.Join(s.baseDir, cleanKey)

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

func safeObjectKey(key string) (string, error) {
	cleanKey := filepath.Clean(strings.TrimSpace(key))
	cleanKey = strings.TrimPrefix(cleanKey, string(filepath.Separator))

	if cleanKey == "." || cleanKey == "" {
		return "", errors.New("invalid storage key")
	}

	if strings.HasPrefix(cleanKey, "..") {
		return "", errors.New("invalid storage key")
	}

	return cleanKey, nil
}
