package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ifaisalabid1/notes-platform-api/internal/admin"
	"github.com/ifaisalabid1/notes-platform-api/internal/chapter"
	"github.com/ifaisalabid1/notes-platform-api/internal/http/handlers"
	"github.com/ifaisalabid1/notes-platform-api/internal/note"
	"github.com/ifaisalabid1/notes-platform-api/internal/semester"
	"github.com/ifaisalabid1/notes-platform-api/internal/subject"
	"github.com/ifaisalabid1/notes-platform-api/internal/unit"
)

type RouterDeps struct {
	Database       handlers.DatabasePinger
	DBPool         *pgxpool.Pool
	Logger         *slog.Logger
	SessionManager *scs.SessionManager
	OwnerEmail     string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestSize(20 << 20))
	r.Use(deps.SessionManager.LoadAndSave)

	healthHandler := handlers.NewHealthHandler(deps.Database, deps.Logger)

	adminRepository := admin.NewRepository(deps.DBPool)
	adminService := admin.NewService(adminRepository, deps.OwnerEmail)
	adminHandler := admin.NewHandler(adminService, deps.SessionManager, deps.Logger)
	adminMiddleware := admin.NewMiddleware(adminService, deps.SessionManager, deps.Logger)

	semesterRepository := semester.NewRepository(deps.DBPool)
	semesterService := semester.NewService(semesterRepository)
	semesterHandler := semester.NewHandler(semesterService, deps.Logger)

	subjectRepository := subject.NewRepository(deps.DBPool)
	subjectService := subject.NewService(subjectRepository)
	subjectHandler := subject.NewHandler(subjectService, deps.Logger)

	unitRepository := unit.NewRepository(deps.DBPool)
	unitService := unit.NewService(unitRepository)
	unitHandler := unit.NewHandler(unitService, deps.Logger)

	chapterRepository := chapter.NewRepository(deps.DBPool)
	chapterService := chapter.NewService(chapterRepository)
	chapterHandler := chapter.NewHandler(chapterService, deps.Logger)

	noteRepository := note.NewRepository(deps.DBPool)
	noteService := note.NewService(noteRepository)
	noteHandler := note.NewHandler(noteService, deps.Logger)

	r.Get("/healthz", healthHandler.Check)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/public", func(r chi.Router) {
			r.Get("/semesters", semesterHandler.ListPublic)
			r.Get("/semesters/{semesterID}", semesterHandler.GetPublicByID)

			r.Get("/semesters/{semesterID}/subjects", subjectHandler.ListPublicBySemester)
			r.Get("/subjects/{subjectID}", subjectHandler.GetPublicByID)

			r.Get("/subjects/{subjectID}/units", unitHandler.ListPublicBySubject)
			r.Get("/units/{unitID}", unitHandler.GetPublicByID)

			r.Get("/units/{unitID}/chapters", chapterHandler.ListPublicByUnit)
			r.Get("/chapters/{chapterID}", chapterHandler.GetPublicByID)

			r.Get("/chapters/{chapterID}/notes", noteHandler.ListPublicByChapter)
			r.Get("/notes/{noteID}", noteHandler.GetPublicByID)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Post("/bootstrap-owner", adminHandler.BootstrapOwner)
			r.Post("/login", adminHandler.Login)

			r.Group(func(r chi.Router) {
				r.Use(adminMiddleware.RequireAdmin)

				r.Post("/logout", adminHandler.Logout)
				r.Get("/me", adminHandler.Me)

				r.With(adminMiddleware.RequireOwner).Post("/admins", adminHandler.CreateAdmin)

				r.Get("/semesters", semesterHandler.ListAdmin)
				r.Post("/semesters", semesterHandler.Create)
				r.Get("/semesters/{semesterID}", semesterHandler.GetAdminByID)
				r.Patch("/semesters/{semesterID}", semesterHandler.Update)
				r.Delete("/semesters/{semesterID}", semesterHandler.Delete)

				r.Get("/semesters/{semesterID}/subjects", subjectHandler.ListAdminBySemester)
				r.Post("/semesters/{semesterID}/subjects", subjectHandler.Create)
				r.Get("/subjects/{subjectID}", subjectHandler.GetAdminByID)
				r.Patch("/subjects/{subjectID}", subjectHandler.Update)
				r.Delete("/subjects/{subjectID}", subjectHandler.Delete)

				r.Get("/subjects/{subjectID}/units", unitHandler.ListAdminBySubject)
				r.Post("/subjects/{subjectID}/units", unitHandler.Create)
				r.Get("/units/{unitID}", unitHandler.GetAdminByID)
				r.Patch("/units/{unitID}", unitHandler.Update)
				r.Delete("/units/{unitID}", unitHandler.Delete)

				r.Get("/units/{unitID}/chapters", chapterHandler.ListAdminByUnit)
				r.Post("/units/{unitID}/chapters", chapterHandler.Create)
				r.Get("/chapters/{chapterID}", chapterHandler.GetAdminByID)
				r.Patch("/chapters/{chapterID}", chapterHandler.Update)
				r.Delete("/chapters/{chapterID}", chapterHandler.Delete)

				r.Get("/chapters/{chapterID}/notes", noteHandler.ListAdminByChapter)
				r.Post("/chapters/{chapterID}/notes", noteHandler.Create)
				r.Get("/notes/{noteID}", noteHandler.GetAdminByID)
				r.Patch("/notes/{noteID}", noteHandler.Update)
				r.Delete("/notes/{noteID}", noteHandler.Delete)
			})
		})
	})

	return r
}
