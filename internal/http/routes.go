package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ifaisalabid1/notes-platform-api/internal/http/handlers"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	healthHandler := handlers.NewHealthHandler()

	r.Get("/healthz", healthHandler.Check)

	return r
}
