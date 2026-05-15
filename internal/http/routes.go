package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ifaisalabid1/notes-platform-api/internal/http/handlers"
)

type RouterDeps struct {
	Database handlers.DatabasePinger
	Logger   *slog.Logger
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	healthHandler := handlers.NewHealthHandler(deps.Database, deps.Logger)

	r.Get("/healthz", healthHandler.Check)

	return r
}
