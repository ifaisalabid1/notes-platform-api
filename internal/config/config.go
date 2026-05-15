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
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		OwnerEmail:        os.Getenv("OWNER_EMAIL"),
		SessionCookieName: getEnv("SESSION_COOKIE_NAME", "notes_platform_session"),
		CookieDomain:      os.Getenv("COOKIE_DOMAIN"),
	}

	cookieSecure, err := strconv.ParseBool(getEnv("COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, errors.New("COOKIE_SECURE must be true or false")
	}
	cfg.CookieSecure = cookieSecure

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if cfg.OwnerEmail == "" {
		return Config{}, errors.New("OWNER_EMAIL is required")
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
