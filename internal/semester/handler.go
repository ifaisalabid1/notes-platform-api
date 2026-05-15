package semester

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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
		writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	semester, err := h.service.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, ErrTitleRequired):
			writeError(w, http.StatusBadRequest, "title_required", "Title is required.")
			return
		case errors.Is(err, ErrSlugRequired):
			writeError(w, http.StatusBadRequest, "slug_required", "Slug is required.")
			return
		case errors.Is(err, ErrInvalidSlug):
			writeError(w, http.StatusBadRequest, "invalid_slug", "Slug may only contain lowercase letters, numbers, and hyphens.")
			return
		case errors.Is(err, ErrSemesterSlugConflicts):
			writeError(w, http.StatusConflict, "semester_slug_conflict", "A semester with this slug already exists.")
			return
		default:
			h.logger.Error("failed to create semester", slog.Any("error", err))
			writeError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
			return
		}
	}

	writeJSON(w, http.StatusCreated, semester)
}

func (h *Handler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	semesters, err := h.service.ListAdmin(r.Context())
	if err != nil {
		h.logger.Error("failed to list admin semesters", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": semesters,
	})
}

func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	semesters, err := h.service.ListPublic(r.Context())
	if err != nil {
		h.logger.Error("failed to list public semesters", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": semesters,
	})
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(w, statusCode, errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}
