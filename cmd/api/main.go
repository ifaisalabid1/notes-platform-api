package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/joho/godotenv"

	"github.com/ifaisalabid1/notes-platform-api/internal/config"
	apphttp "github.com/ifaisalabid1/notes-platform-api/internal/http"
	"github.com/ifaisalabid1/notes-platform-api/internal/platform/database"
	"github.com/ifaisalabid1/notes-platform-api/internal/platform/logger"
	"github.com/ifaisalabid1/notes-platform-api/internal/storage"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logr := logger.New(cfg.AppEnv)

	rootCtx := context.Background()

	db, err := database.New(rootCtx, cfg.DatabaseURL, logr)
	if err != nil {
		logr.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(db.Pool)
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.IdleTimeout = 2 * time.Hour
	sessionManager.Cookie.Name = cfg.SessionCookieName
	sessionManager.Cookie.HttpOnly = true
	sessionManager.Cookie.Persist = true
	sessionManager.Cookie.SameSite = http.SameSiteLaxMode
	sessionManager.Cookie.Secure = cfg.CookieSecure

	if cfg.CookieDomain != "" {
		sessionManager.Cookie.Domain = cfg.CookieDomain
	}

	objectStorage, err := storage.NewFromConfig(rootCtx, storage.FactoryConfig{
		Driver:            cfg.StorageDriver,
		LocalStorageDir:   cfg.LocalStorageDir,
		R2AccountID:       cfg.R2AccountID,
		R2AccessKeyID:     cfg.R2AccessKeyID,
		R2SecretAccessKey: cfg.R2SecretAccessKey,
		R2BucketName:      cfg.R2BucketName,
	})
	if err != nil {
		logr.Error("failed to initialize object storage", slog.Any("error", err))
		os.Exit(1)
	}

	router := apphttp.NewRouter(apphttp.RouterDeps{
		Database:          db,
		DBPool:            db.Pool,
		Logger:            logr,
		SessionManager:    sessionManager,
		OwnerEmail:        cfg.OwnerEmail,
		ObjectStorage:     objectStorage,
		UploadMaxBytes:    cfg.UploadMaxBytes,
		PublicFileBaseURL: cfg.PublicFileBaseURL,
		WorkerAPISecret:   cfg.WorkerAPISecret,
	})

	server := apphttp.NewServer(cfg.HTTPPort, router, logr)

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			logr.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	case sig := <-quit:
		logr.Info("received shutdown signal", slog.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logr.Error("graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	logr.Info("server stopped")
}
