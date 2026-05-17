package admin

import (
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/audit"
	"github.com/ifaisalabid1/notes-platform-api/internal/http/response"
)

type Middleware struct {
	service        *Service
	sessionManager *scs.SessionManager
	logger         *slog.Logger
}

func NewMiddleware(service *Service, sessionManager *scs.SessionManager, logger *slog.Logger) *Middleware {
	return &Middleware{
		service:        service,
		sessionManager: sessionManager,
		logger:         logger,
	}
}

func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawAdminID := m.sessionManager.GetString(r.Context(), "admin_id")
		if rawAdminID == "" {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
			return
		}

		adminID, err := uuid.Parse(rawAdminID)
		if err != nil {
			_ = m.sessionManager.Destroy(r.Context())
			response.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid session.")
			return
		}

		currentAdmin, err := m.service.GetByID(r.Context(), adminID)
		if err != nil {
			m.logger.Error("failed to load admin from session", slog.Any("error", err))
			_ = m.sessionManager.Destroy(r.Context())
			response.Error(w, http.StatusUnauthorized, "unauthorized", "Invalid session.")
			return
		}

		if !currentAdmin.IsActive {
			_ = m.sessionManager.Destroy(r.Context())
			response.Error(w, http.StatusForbidden, "admin_inactive", "Admin account is inactive.")
			return
		}

		ctx := ContextWithAdmin(r.Context(), currentAdmin)
		ctx = audit.ContextWithActorID(ctx, currentAdmin.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentAdmin, ok := CurrentAdmin(r.Context())
		if !ok {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
			return
		}

		if currentAdmin.Role != RoleOwner {
			response.Error(w, http.StatusForbidden, "forbidden", "Only the owner admin can perform this action.")
			return
		}

		next.ServeHTTP(w, r)
	})
}
