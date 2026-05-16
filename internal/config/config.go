package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	AppEnv      string
	HTTPPort    string
	DatabaseURL string

	OwnerEmail        string
	SessionCookieName string
	CookieSecure      bool
	CookieDomain      string

	UploadMaxBytes  int64
	LocalStorageDir string

	StorageDriver     string
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string

	WorkerAPISecret   string
	PublicFileBaseURL string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		OwnerEmail:        os.Getenv("OWNER_EMAIL"),
		SessionCookieName: getEnv("SESSION_COOKIE_NAME", "notes_platform_session"),
		CookieDomain:      os.Getenv("COOKIE_DOMAIN"),
		LocalStorageDir:   getEnv("LOCAL_STORAGE_DIR", "./storage"),

		StorageDriver:     getEnv("STORAGE_DRIVER", "local"),
		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),

		WorkerAPISecret:   os.Getenv("WORKER_API_SECRET"),
		PublicFileBaseURL: getEnv("PUBLIC_FILE_BASE_URL", "http://localhost:8787"),
	}

	cookieSecure, err := strconv.ParseBool(getEnv("COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, errors.New("COOKIE_SECURE must be true or false")
	}
	cfg.CookieSecure = cookieSecure

	uploadMaxBytes, err := strconv.ParseInt(getEnv("UPLOAD_MAX_BYTES", "52428800"), 10, 64)
	if err != nil || uploadMaxBytes <= 0 {
		return Config{}, errors.New("UPLOAD_MAX_BYTES must be a positive integer")
	}
	cfg.UploadMaxBytes = uploadMaxBytes

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.OwnerEmail == "" {
		return Config{}, errors.New("OWNER_EMAIL is required")
	}

	if cfg.WorkerAPISecret == "" {
		return Config{}, errors.New("WORKER_API_SECRET is required")
	}

	switch cfg.StorageDriver {
	case "local":
	case "r2":
		if cfg.R2AccountID == "" {
			return Config{}, errors.New("R2_ACCOUNT_ID is required when STORAGE_DRIVER=r2")
		}

		if cfg.R2AccessKeyID == "" {
			return Config{}, errors.New("R2_ACCESS_KEY_ID is required when STORAGE_DRIVER=r2")
		}

		if cfg.R2SecretAccessKey == "" {
			return Config{}, errors.New("R2_SECRET_ACCESS_KEY is required when STORAGE_DRIVER=r2")
		}

		if cfg.R2BucketName == "" {
			return Config{}, errors.New("R2_BUCKET_NAME is required when STORAGE_DRIVER=r2")
		}
	default:
		return Config{}, errors.New("STORAGE_DRIVER must be either local or r2")
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
