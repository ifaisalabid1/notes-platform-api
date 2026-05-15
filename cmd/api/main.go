package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/ifaisalabid1/notes-platform-api/internal/config"
	apphttp "github.com/ifaisalabid1/notes-platform-api/internal/http"
	"github.com/ifaisalabid1/notes-platform-api/internal/platform/database"
	"github.com/ifaisalabid1/notes-platform-api/internal/platform/logger"
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

	router := apphttp.NewRouter(apphttp.RouterDeps{
		Database: db,
		DBPool:   db.Pool,
		Logger:   logr,
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
