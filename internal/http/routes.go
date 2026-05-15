package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ifaisalabid1/notes-platform-api/internal/http/handlers"
	"github.com/ifaisalabid1/notes-platform-api/internal/semester"
)

type RouterDeps struct {
	Database handlers.DatabasePinger
	DBPool   *pgxpool.Pool
	Logger   *slog.Logger
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestSize(20 << 20))

	healthHandler := handlers.NewHealthHandler(deps.Database, deps.Logger)

	semesterRepository := semester.NewRepository(deps.DBPool)
	semesterService := semester.NewService(semesterRepository)
	semesterHandler := semester.NewHandler(semesterService, deps.Logger)

	r.Get("/healthz", healthHandler.Check)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/public", func(r chi.Router) {
			r.Get("/semesters", semesterHandler.ListPublic)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Get("/semesters", semesterHandler.ListAdmin)
			r.Post("/semesters", semesterHandler.Create)
		})
	})

	return r
}
