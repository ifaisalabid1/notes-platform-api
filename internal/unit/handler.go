package unit

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/http/response"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	var input CreateUnitInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	unit, err := h.service.Create(r.Context(), subjectID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to create unit")
		return
	}

	response.JSON(w, http.StatusCreated, unit)
}

func (h *Handler) ListAdminBySubject(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	units, err := h.service.ListAdminBySubject(r.Context(), subjectID)
	if err != nil {
		h.logger.Error("failed to list admin units", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": units,
	})
}

func (h *Handler) ListPublicBySubject(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	units, err := h.service.ListPublicBySubject(r.Context(), subjectID)
	if err != nil {
		h.logger.Error("failed to list public units", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": units,
	})
}

func (h *Handler) GetAdminByID(w http.ResponseWriter, r *http.Request) {
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	unit, err := h.service.GetAdminByID(r.Context(), unitID)
	if err != nil {
		h.handleReadError(w, err, "failed to get admin unit")
		return
	}

	response.JSON(w, http.StatusOK, unit)
}

func (h *Handler) GetPublicByID(w http.ResponseWriter, r *http.Request) {
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	unit, err := h.service.GetPublicByID(r.Context(), unitID)
	if err != nil {
		h.handleReadError(w, err, "failed to get public unit")
		return
	}

	response.JSON(w, http.StatusOK, unit)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	var input UpdateUnitInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	unit, err := h.service.Update(r.Context(), unitID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to update unit")
		return
	}

	response.JSON(w, http.StatusOK, unit)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), unitID); err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			response.Error(w, http.StatusNotFound, "unit_not_found", "Unit was not found.")
			return
		}

		h.logger.Error("failed to delete unit", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReadError(w http.ResponseWriter, err error, logMessage string) {
	if errors.Is(err, ErrUnitNotFound) {
		response.Error(w, http.StatusNotFound, "unit_not_found", "Unit was not found.")
		return
	}

	h.logger.Error(logMessage, slog.Any("error", err))
	response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
}

func (h *Handler) handleWriteError(w http.ResponseWriter, err error, logMessage string) {
	switch {
	case errors.Is(err, ErrTitleRequired):
		response.Error(w, http.StatusBadRequest, "title_required", "Title is required.")
	case errors.Is(err, ErrSlugRequired):
		response.Error(w, http.StatusBadRequest, "slug_required", "Slug is required.")
	case errors.Is(err, ErrInvalidSlug):
		response.Error(w, http.StatusBadRequest, "invalid_slug", "Slug may only contain lowercase letters, numbers, and hyphens.")
	case errors.Is(err, ErrUnitSlugConflicts):
		response.Error(w, http.StatusConflict, "unit_slug_conflict", "A unit with this slug already exists in this subject.")
	case errors.Is(err, ErrSubjectNotFound):
		response.Error(w, http.StatusNotFound, "subject_not_found", "Subject was not found.")
	case errors.Is(err, ErrUnitNotFound):
		response.Error(w, http.StatusNotFound, "unit_not_found", "Unit was not found.")
	default:
		h.logger.Error(logMessage, slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
	}
}

func parseUUIDParam(
	w http.ResponseWriter,
	r *http.Request,
	paramName string,
	errorCode string,
	errorMessage string,
) (uuid.UUID, bool) {
	rawID := chi.URLParam(r, paramName)

	id, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errorCode, errorMessage)
		return uuid.Nil, false
	}

	return id, true
}
