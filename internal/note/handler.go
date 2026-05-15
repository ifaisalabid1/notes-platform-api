package note

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/http/response"
)

type Handler struct {
	service        *Service
	logger         *slog.Logger
	uploadMaxBytes int64
}

func NewHandler(service *Service, logger *slog.Logger, uploadMaxBytes int64) *Handler {
	return &Handler{
		service:        service,
		logger:         logger,
		uploadMaxBytes: uploadMaxBytes,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	var input CreateNoteInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	note, err := h.service.Create(r.Context(), chapterID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to create note")
		return
	}

	response.JSON(w, http.StatusCreated, note)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.uploadMaxBytes)

	if err := r.ParseMultipartForm(h.uploadMaxBytes); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_multipart_form", "Request must be multipart/form-data and within the upload size limit.")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file_required", "File is required.")
		return
	}
	defer file.Close()

	description := nullableFormValue(r, "description")

	sortOrder, err := parseIntFormValue(r.FormValue("sort_order"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_sort_order", "Sort order must be a valid integer.")
		return
	}

	input := UploadNoteInput{
		Title:       r.FormValue("title"),
		Slug:        r.FormValue("slug"),
		Description: description,
		SortOrder:   sortOrder,
		IsPublished: ParseBoolFormValue(r.FormValue("is_published")),
		File:        file,
		FileHeader:  fileHeader,
	}

	createdNote, err := h.service.Upload(r.Context(), chapterID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to upload note")
		return
	}

	response.JSON(w, http.StatusCreated, createdNote)
}

func (h *Handler) ListAdminByChapter(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	notes, err := h.service.ListAdminByChapter(r.Context(), chapterID)
	if err != nil {
		h.logger.Error("failed to list admin notes", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": notes,
	})
}

func (h *Handler) ListPublicByChapter(w http.ResponseWriter, r *http.Request) {
	chapterID, ok := parseUUIDParam(w, r, "chapterID", "invalid_chapter_id", "Chapter ID must be a valid UUID.")
	if !ok {
		return
	}

	notes, err := h.service.ListPublicByChapter(r.Context(), chapterID)
	if err != nil {
		h.logger.Error("failed to list public notes", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"data": notes,
	})
}

func (h *Handler) GetAdminByID(w http.ResponseWriter, r *http.Request) {
	noteID, ok := parseUUIDParam(w, r, "noteID", "invalid_note_id", "Note ID must be a valid UUID.")
	if !ok {
		return
	}

	note, err := h.service.GetAdminByID(r.Context(), noteID)
	if err != nil {
		h.handleReadError(w, err, "failed to get admin note")
		return
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *Handler) GetPublicByID(w http.ResponseWriter, r *http.Request) {
	noteID, ok := parseUUIDParam(w, r, "noteID", "invalid_note_id", "Note ID must be a valid UUID.")
	if !ok {
		return
	}

	note, err := h.service.GetPublicByID(r.Context(), noteID)
	if err != nil {
		h.handleReadError(w, err, "failed to get public note")
		return
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	noteID, ok := parseUUIDParam(w, r, "noteID", "invalid_note_id", "Note ID must be a valid UUID.")
	if !ok {
		return
	}

	var input UpdateNoteInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	note, err := h.service.Update(r.Context(), noteID, input)
	if err != nil {
		h.handleWriteError(w, err, "failed to update note")
		return
	}

	response.JSON(w, http.StatusOK, note)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	noteID, ok := parseUUIDParam(w, r, "noteID", "invalid_note_id", "Note ID must be a valid UUID.")
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), noteID); err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			response.Error(w, http.StatusNotFound, "note_not_found", "Note was not found.")
			return
		}

		h.logger.Error("failed to delete note", slog.Any("error", err))
		response.Error(w, http.StatusInternalServerError, "internal_error", "Something went wrong.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReadError(w http.ResponseWriter, err error, logMessage string) {
	if errors.Is(err, ErrNoteNotFound) {
		response.Error(w, http.StatusNotFound, "note_not_found", "Note was not found.")
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
	case errors.Is(err, ErrOriginalFileNameRequired):
		response.Error(w, http.StatusBadRequest, "original_file_name_required", "Original file name is required.")
	case errors.Is(err, ErrStoredObjectKeyRequired):
		response.Error(w, http.StatusBadRequest, "stored_object_key_required", "Stored object key is required.")
	case errors.Is(err, ErrFileContentTypeRequired):
		response.Error(w, http.StatusBadRequest, "file_content_type_required", "File content type is required.")
	case errors.Is(err, ErrInvalidFileSize):
		response.Error(w, http.StatusBadRequest, "invalid_file_size", "File size must be greater than zero.")
	case errors.Is(err, ErrFileRequired):
		response.Error(w, http.StatusBadRequest, "file_required", "File is required.")
	case errors.Is(err, ErrUnsupportedFileType):
		response.Error(w, http.StatusBadRequest, "unsupported_file_type", "Only PDF, JPEG, PNG, and WebP files are supported.")
	case errors.Is(err, ErrFileTooLarge):
		response.Error(w, http.StatusRequestEntityTooLarge, "file_too_large", "Uploaded file is too large.")
	case errors.Is(err, ErrNoteSlugConflicts):
		response.Error(w, http.StatusConflict, "note_slug_conflict", "A note with this slug already exists in this chapter.")
	case errors.Is(err, ErrObjectKeyConflicts):
		response.Error(w, http.StatusConflict, "object_key_conflict", "A note with this stored object key already exists.")
	case errors.Is(err, ErrChapterNotFound):
		response.Error(w, http.StatusNotFound, "chapter_not_found", "Chapter was not found.")
	case errors.Is(err, ErrNoteNotFound):
		response.Error(w, http.StatusNotFound, "note_not_found", "Note was not found.")
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

func nullableFormValue(r *http.Request, key string) *string {
	value := strings.TrimSpace(r.FormValue(key))
	if value == "" {
		return nil
	}

	return &value
}

func parseIntFormValue(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	return strconv.Atoi(value)
}
