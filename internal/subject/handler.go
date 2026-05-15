package subject

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
	semesterID, ok := parseUUIDParam(w, r, "semesterID", "invalid_semester_id", "Semester ID must be a valid UUID.")
	if !ok {
		return
	}

	var input CreateSubjectInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	subject, err := h.service.Create(r.Context(), semesterID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to create subject")
		return
	}

	response.JSON(w, http.StatusCreated, subject)
}

func (h *Handler) ListAdminBySemester(w http.ResponseWriter, r *http.Request) {
	semesterID, ok := parseUUIDParam(w, r, "semesterID", "invalid_semester_id", "Semester ID must be a valid UUID.")
	if !ok {
		return
	}

	subjects, err := h.service.ListAdminBySemester(r.Context(), semesterID)
	if err != nil {
		h.logger.Error("failed to list admin subjects", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": subjects,
	})
}

func (h *Handler) ListPublicBySemester(w http.ResponseWriter, r *http.Request) {
	semesterID, ok := parseUUIDParam(w, r, "semesterID", "invalid_semester_id", "Semester ID must be a valid UUID.")
	if !ok {
		return
	}

	subjects, err := h.service.ListPublicBySemester(r.Context(), semesterID)
	if err != nil {
		h.logger.Error("failed to list public subjects", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": subjects,
	})
}

func (h *Handler) GetAdminByID(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	subject, err := h.service.GetAdminByID(r.Context(), subjectID)
	if err != nil {
		h.handleReadError(w, err, "failed to get admin subject")
		return
	}

	response.JSON(w, http.StatusOK, subject)
}

func (h *Handler) GetPublicByID(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	subject, err := h.service.GetPublicByID(r.Context(), subjectID)
	if err != nil {
		h.handleReadError(w, err, "failed to get public subject")
		return
	}

	response.JSON(w, http.StatusOK, subject)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	var input UpdateSubjectInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	subject, err := h.service.Update(r.Context(), subjectID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to update subject")
		return
	}

	response.JSON(w, http.StatusOK, subject)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	subjectID, ok := parseUUIDParam(w, r, "subjectID", "invalid_subject_id", "Subject ID must be a valid UUID.")
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), subjectID); err != nil {
		if errors.Is(err, ErrSubjectNotFound) {
			response.Error(w, http.StatusNotFound, "subject_not_found", "Subject was not found.")
			return
		}

		h.logger.Error("failed to delete subject", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReadError(w http.ResponseWriter, err error, logMessage string) {
	if errors.Is(err, ErrSubjectNotFound) {
		response.Error(w, http.StatusNotFound, "subject_not_found", "Subject was not found.")
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
	case errors.Is(err, ErrSubjectSlugConflicts):
		response.Error(w, http.StatusConflict, "subject_slug_conflict", "A subject with this slug already exists in this semester.")
	case errors.Is(err, ErrSemesterNotFound):
		response.Error(w, http.StatusNotFound, "semester_not_found", "Semester was not found.")
	case errors.Is(err, ErrSubjectNotFound):
		response.Error(w, http.StatusNotFound, "subject_not_found", "Subject was not found.")
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
