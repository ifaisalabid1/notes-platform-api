package semester

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
	var input CreateSemesterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	semester, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.handleWriteError(w, err, "failed to create semester")
		return
	}

	response.JSON(w, http.StatusCreated, semester)
}

func (h *Handler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	semesters, err := h.service.ListAdmin(r.Context())
	if err != nil {
		h.logger.Error("failed to list admin semesters", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": semesters,
	})
}

func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	semesters, err := h.service.ListPublic(r.Context())
	if err != nil {
		h.logger.Error("failed to list public semesters", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": semesters,
	})
}

func (h *Handler) GetAdminByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseSemesterID(w, r)
	if !ok {
		return
	}

	semester, err := h.service.GetAdminByID(r.Context(), id)
	if err != nil {
		h.handleReadError(w, err, "failed to get admin semester")
		return
	}

	response.JSON(w, http.StatusOK, semester)
}

func (h *Handler) GetPublicByID(w http.ResponseWriter, r *http.Request) {
	id, ok := parseSemesterID(w, r)
	if !ok {
		return
	}

	semester, err := h.service.GetPublicByID(r.Context(), id)
	if err != nil {
		h.handleReadError(w, err, "failed to get public semester")
		return
	}

	response.JSON(w, http.StatusOK, semester)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseSemesterID(w, r)
	if !ok {
		return
	}

	var input UpdateSemesterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	semester, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to update semester")
		return
	}

	response.JSON(w, http.StatusOK, semester)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseSemesterID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrSemesterNotFound) {
			response.Error(w, http.StatusNotFound, "semester_not_found", "Semester was not found.")
			return
		}

		h.logger.Error("failed to delete semester", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReadError(w http.ResponseWriter, err error, logMessage string) {
	if errors.Is(err, ErrSemesterNotFound) {
		response.Error(w, http.StatusNotFound, "semester_not_found", "Semester was not found.")
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
	case errors.Is(err, ErrSemesterSlugConflicts):
		response.Error(w, http.StatusConflict, "semester_slug_conflict", "A semester with this slug already exists.")
	case errors.Is(err, ErrSemesterNotFound):
		response.Error(w, http.StatusNotFound, "semester_not_found", "Semester was not found.")
	default:
		h.logger.Error(logMessage, slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
	}
}

func parseSemesterID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	rawID := chi.URLParam(r, "semesterID")

	id, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_semester_id", "Semester ID must be a valid UUID.")
		return uuid.Nil, false
	}

	return id, true
}
