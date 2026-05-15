package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type DatabasePinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	database DatabasePinger
	logger   *slog.Logger
}

func NewHealthHandler(database DatabasePinger, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		database: database,
		logger:   logger,
	}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	response := map[string]string{
		"status":   "ok",
		"database": "ok",
	}

	if err := h.database.Ping(ctx); err != nil {
		h.logger.Error("database health check failed", slog.Any("error", err))

		response["status"] = "degraded"
		response["database"] = "error"

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(response)
}
