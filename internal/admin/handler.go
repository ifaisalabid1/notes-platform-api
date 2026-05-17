package admin

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/http/response"
)

type Handler struct {
	service        *Service
	sessionManager *scs.SessionManager
	logger         *slog.Logger
}

func NewHandler(service *Service, sessionManager *scs.SessionManager, logger *slog.Logger) *Handler {
	return &Handler{
		service:        service,
		sessionManager: sessionManager,
		logger:         logger,
	}
}

func (h *Handler) BootstrapOwner(w http.ResponseWriter, r *http.Request) {
	var input BootstrapOwnerInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	createdAdmin, err := h.service.BootstrapOwner(r.Context(), input)
	if err != nil {
		h.handleAuthWriteError(w, err, "failed to bootstrap owner")
		return
	}

	if err := h.sessionManager.RenewToken(r.Context()); err != nil {
		h.logger.Error("failed to renew session token", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	h.sessionManager.Put(r.Context(), "admin_id", createdAdmin.ID.String())
	h.sessionManager.Put(r.Context(), "admin_role", string(createdAdmin.Role))

	response.JSON(w, http.StatusCreated, createdAdmin)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	loggedInAdmin, err := h.service.Login(r.Context(), input)
	if err != nil {
		h.handleLoginError(w, err)
		return
	}

	if err := h.sessionManager.RenewToken(r.Context()); err != nil {
		h.logger.Error("failed to renew session token", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	h.sessionManager.Put(r.Context(), "admin_id", loggedInAdmin.ID.String())
	h.sessionManager.Put(r.Context(), "admin_role", string(loggedInAdmin.Role))

	response.JSON(w, http.StatusOK, loggedInAdmin)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionManager.Destroy(r.Context()); err != nil {
		h.logger.Error("failed to destroy session", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	currentAdmin, ok := CurrentAdmin(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
		return
	}

	response.JSON(w, http.StatusOK, currentAdmin)
}

func (h *Handler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	currentAdmin, ok := CurrentAdmin(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
		return
	}

	var input CreateAdminInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	createdAdmin, err := h.service.CreateAdmin(r.Context(), currentAdmin, input)
	if err != nil {
		h.handleAuthWriteError(w, err, "failed to create admin")
		return
	}

	response.JSON(w, http.StatusCreated, createdAdmin)
}

func (h *Handler) handleLoginError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password.")
	case errors.Is(err, ErrInactiveAdmin):
		response.Error(w, http.StatusForbidden, "admin_inactive", "Admin account is inactive.")
	default:
		h.logger.Error("failed to login", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
	}
}

func (h *Handler) handleAuthWriteError(w http.ResponseWriter, err error, logMessage string) {
	switch {
	case errors.Is(err, ErrEmailRequired):
		response.Error(w, http.StatusBadRequest, "email_required", "Email is required.")
	case errors.Is(err, ErrInvalidEmail):
		response.Error(w, http.StatusBadRequest, "invalid_email", "Email is invalid.")
	case errors.Is(err, ErrPasswordRequired):
		response.Error(w, http.StatusBadRequest, "password_required", "Password is required.")
	case errors.Is(err, ErrPasswordTooShort):
		response.Error(w, http.StatusBadRequest, "password_too_short", "Password must be at least 8 characters.")
	case errors.Is(err, ErrDisplayNameRequired):
		response.Error(w, http.StatusBadRequest, "display_name_required", "Display name is required.")
	case errors.Is(err, ErrEmailConflict):
		response.Error(w, http.StatusConflict, "email_conflict", "An admin with this email already exists.")
	case errors.Is(err, ErrOwnerExists):
		response.Error(w, http.StatusConflict, "owner_exists", "Owner admin already exists.")
	case errors.Is(err, ErrOwnerNotAllowed):
		response.Error(w, http.StatusForbidden, "owner_not_allowed", "This email is not allowed to become owner.")
	case errors.Is(err, ErrForbidden):
		response.Error(w, http.StatusForbidden, "forbidden", "You are not allowed to perform this action.")
	default:
		h.logger.Error(logMessage, slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
	}
}

func (h *Handler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	currentAdmin, ok := CurrentAdmin(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
		return
	}

	admins, err := h.service.ListAdmins(r.Context(), currentAdmin)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.Error(w, http.StatusForbidden, "forbidden", "Only the owner admin can perform this action.")
			return
		}

		h.logger.Error("failed to list admins", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": admins,
	})
}

func (h *Handler) UpdateAdminStatus(w http.ResponseWriter, r *http.Request) {
	currentAdmin, ok := CurrentAdmin(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
		return
	}

	adminID, ok := parseAdminID(w, r)
	if !ok {
		return
	}

	var input UpdateAdminStatusInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	updatedAdmin, err := h.service.UpdateAdminStatus(r.Context(), currentAdmin, adminID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			response.Error(w, http.StatusForbidden, "forbidden", "Only the owner admin can perform this action.")
		case errors.Is(err, ErrCannotDeactivateSelf):
			response.Error(w, http.StatusBadRequest, "cannot_deactivate_self", "The owner admin cannot be deactivated.")
		case errors.Is(err, ErrAdminNotFound):
			response.Error(w, http.StatusNotFound, "admin_not_found", "Admin was not found.")
		default:
			h.logger.Error("failed to update admin status", slog.Any("error", err))
			response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		}

		return
	}

	response.JSON(w, http.StatusOK, updatedAdmin)
}

func parseAdminID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	rawID := chi.URLParam(r, "adminID")

	id, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_admin_id", "Admin ID must be a valid UUID.")
		return uuid.Nil, false
	}

	return id, true
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	currentAdmin, ok := CurrentAdmin(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
		return
	}

	var input ChangePasswordInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	if err := h.service.ChangePassword(r.Context(), currentAdmin, input); err != nil {
		switch {
		case errors.Is(err, ErrCurrentPasswordRequired):
			response.Error(w, http.StatusBadRequest, "current_password_required", "Current password is required.")
		case errors.Is(err, ErrNewPasswordRequired):
			response.Error(w, http.StatusBadRequest, "new_password_required", "New password is required.")
		case errors.Is(err, ErrPasswordTooShort):
			response.Error(w, http.StatusBadRequest, "password_too_short", "Password must be at least 10 characters.")
		case errors.Is(err, ErrSamePassword):
			response.Error(w, http.StatusBadRequest, "same_password", "New password must be different from current password.")
		case errors.Is(err, ErrInvalidCredentials):
			response.Error(w, http.StatusUnauthorized, "invalid_credentials", "Current password is incorrect.")
		case errors.Is(err, ErrInactiveAdmin):
			response.Error(w, http.StatusForbidden, "admin_inactive", "Admin account is inactive.")
		default:
			h.logger.Error("failed to change password", slog.Any("error", err))
			response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		}

		return
	}

	if err := h.sessionManager.RenewToken(r.Context()); err != nil {
		h.logger.Error("failed to renew session token after password change", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	h.sessionManager.Put(r.Context(), "admin_id", currentAdmin.ID.String())
	h.sessionManager.Put(r.Context(), "admin_role", string(currentAdmin.Role))

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Password changed successfully.",
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	currentAdmin, ok := CurrentAdmin(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "unauthorized", "You must be logged in.")
		return
	}

	var input UpdateProfileInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	updatedAdmin, err := h.service.UpdateProfile(r.Context(), currentAdmin, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrDisplayNameRequired):
			response.Error(w, http.StatusBadRequest, "display_name_required", "Display name is required.")
		case errors.Is(err, ErrAdminNotFound):
			response.Error(w, http.StatusNotFound, "admin_not_found", "Admin was not found.")
		default:
			h.logger.Error("failed to update admin profile", slog.Any("error", err))
			response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		}

		return
	}

	response.JSON(w, http.StatusOK, updatedAdmin)
}
