package chapter

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
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	var input CreateChapterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	chapter, err := h.service.Create(r.Context(), unitID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to create chapter")
		return
	}

	response.JSON(w, http.StatusCreated, chapter)
}

func (h *Handler) ListAdminByUnit(w http.ResponseWriter, r *http.Request) {
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	chapters, err := h.service.ListAdminByUnit(r.Context(), unitID)
	if err != nil {
		h.logger.Error("failed to list admin chapters", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": chapters,
	})
}

func (h *Handler) ListPublicByUnit(w http.ResponseWriter, r *http.Request) {
	unitID, ok := parseUUIDParam(w, r, "unitID", "invalid_unit_id", "Unit ID must be a valid UUID.")
	if !ok {
		return
	}

	chapters, err := h.service.ListPublicByUnit(r.Context(), unitID)
	if err != nil {
		h.logger.Error("failed to list public chapters", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": chapters,
	})
}

func (h *Handler) GetAdminByID(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	chapter, err := h.service.GetAdminByID(r.Context(), chapterID)
	if err != nil {
		h.handleReadError(w, err, "failed to get admin chapter")
		return
	}

	response.JSON(w, http.StatusOK, chapter)
}

func (h *Handler) GetPublicByID(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	chapter, err := h.service.GetPublicByID(r.Context(), chapterID)
	if err != nil {
		h.handleReadError(w, err, "failed to get public chapter")
		return
	}

	response.JSON(w, http.StatusOK, chapter)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	var input UpdateChapterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	chapter, err := h.service.Update(r.Context(), chapterID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to update chapter")
		return
	}

	response.JSON(w, http.StatusOK, chapter)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), chapterID); err != nil {
		if errors.Is(err, ErrChapterNotFound) {
			response.Error(w, http.StatusNotFound, "chapter_not_found", "Chapter was not found.")
			return
		}

		h.logger.Error("failed to delete chapter", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReadError(w http.ResponseWriter, err error, logMessage string) {
	if errors.Is(err, ErrChapterNotFound) {
		response.Error(w, http.StatusNotFound, "chapter_not_found", "Chapter was not found.")
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
	case errors.Is(err, ErrChapterSlugConflicts):
		response.Error(w, http.StatusConflict, "chapter_slug_conflict", "A chapter with this slug already exists in this unit.")
	case errors.Is(err, ErrUnitNotFound):
		response.Error(w, http.StatusNotFound, "unit_not_found", "Unit was not found.")
	case errors.Is(err, ErrChapterNotFound):
		response.Error(w, http.StatusNotFound, "chapter_not_found", "Chapter was not found.")
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
