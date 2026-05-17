package note

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/pagination"
	"github.com/ifaisalabid1/notes-platform-api/internal/storage"
	"github.com/ifaisalabid1/notes-platform-api/internal/validation"
	"github.com/ifaisalabid1/notes-platform-api/internal/watermark"
)

var (
	ErrTitleRequired            = validation.ErrTitleRequired
	ErrSlugRequired             = validation.ErrSlugRequired
	ErrInvalidSlug              = validation.ErrInvalidSlug
	ErrOriginalFileNameRequired = errors.New("original file name is required")
	ErrStoredObjectKeyRequired  = errors.New("stored object key is required")
	ErrFileContentTypeRequired  = errors.New("file content type is required")
	ErrInvalidFileSize          = errors.New("file size must be greater than zero")
	ErrFileRequired             = errors.New("file is required")
	ErrUnsupportedFileType      = errors.New("unsupported file type")
	ErrFileTooLarge             = errors.New("file is too large")
)

type Service struct {
	repository         *Repository
	objectStorage      storage.ObjectStorage
	watermarkProcessor watermark.Processor
	uploadMaxBytes     int64
	publicFileBaseURL  string
}

func NewService(
	repository *Repository,
	objectStorage storage.ObjectStorage,
	watermarkProcessor watermark.Processor,
	uploadMaxBytes int64,
	publicFileBaseURL string,
) *Service {
	return &Service{
		repository:         repository,
		objectStorage:      objectStorage,
		watermarkProcessor: watermarkProcessor,
		uploadMaxBytes:     uploadMaxBytes,
		publicFileBaseURL:  strings.TrimRight(publicFileBaseURL, "/"),
	}
}

func (s *Service) Create(ctx context.Context, chapterID uuid.UUID, input CreateNoteInput) (Note, error) {
	normalized := validation.NormalizeTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	})

	input.Title = normalized.Title
	input.Slug = normalized.Slug

	input.OriginalFileName = strings.TrimSpace(input.OriginalFileName)
	input.StoredObjectKey = strings.TrimSpace(input.StoredObjectKey)
	input.FileContentType = strings.TrimSpace(input.FileContentType)

	if err := validateTitleAndSlug(input.Title, input.Slug); err != nil {
		return Note{}, err
	}

	if input.OriginalFileName == "" {
		return Note{}, ErrOriginalFileNameRequired
	}

	if input.StoredObjectKey == "" {
		return Note{}, ErrStoredObjectKeyRequired
	}

	if input.FileContentType == "" {
		return Note{}, ErrFileContentTypeRequired
	}

	if input.FileSizeBytes <= 0 {
		return Note{}, ErrInvalidFileSize
	}

	return s.repository.Create(ctx, chapterID, input)
}

func (s *Service) Upload(ctx context.Context, chapterID uuid.UUID, input UploadNoteInput) (Note, error) {
	normalized := validation.NormalizeTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	})

	input.Title = normalized.Title
	input.Slug = normalized.Slug

	if err := validateTitleAndSlug(input.Title, input.Slug); err != nil {
		return Note{}, err
	}

	if input.File == nil || input.FileHeader == nil {
		return Note{}, ErrFileRequired
	}

	if input.FileHeader.Size <= 0 {
		return Note{}, ErrInvalidFileSize
	}

	if input.FileHeader.Size > s.uploadMaxBytes {
		return Note{}, ErrFileTooLarge
	}

	contentType, err := detectContentType(input.File)
	if err != nil {
		return Note{}, err
	}

	if !isAllowedContentType(contentType) {
		return Note{}, ErrUnsupportedFileType
	}

	if _, err := input.File.Seek(0, io.SeekStart); err != nil {
		return Note{}, fmt.Errorf("rewind uploaded file: %w", err)
	}

	processedFile, err := s.watermarkProcessor.Process(ctx, watermark.ProcessInput{
		FileName:    input.FileHeader.Filename,
		ContentType: contentType,
		Body:        input.File,
	})
	if err != nil {
		return Note{}, err
	}
	defer processedFile.Cleanup()

	if _, err := processedFile.Body.Seek(0, io.SeekStart); err != nil {
		return Note{}, fmt.Errorf("rewind processed file: %w", err)
	}

	objectKey := buildObjectKey(chapterID, input.FileHeader.Filename)

	putResult, err := s.objectStorage.PutObject(ctx, storage.PutObjectInput{
		Key:         objectKey,
		Body:        processedFile.Body,
		ContentType: processedFile.ContentType,
	})
	if err != nil {
		return Note{}, fmt.Errorf("store uploaded file: %w", err)
	}

	createInput := CreateNoteInput{
		Title:            input.Title,
		Slug:             input.Slug,
		Description:      input.Description,
		OriginalFileName: filepath.Base(input.FileHeader.Filename),
		StoredObjectKey:  putResult.Key,
		FileContentType:  processedFile.ContentType,
		FileSizeBytes:    putResult.SizeBytes,
		IsWatermarked:    processedFile.IsWatermarked,
		IsPublished:      input.IsPublished,
		SortOrder:        input.SortOrder,
	}

	createdNote, err := s.repository.Create(ctx, chapterID, createInput)
	if err != nil {
		_ = s.objectStorage.DeleteObject(ctx, putResult.Key)
		return Note{}, err
	}

	return createdNote, nil
}

func (s *Service) ListAdminByChapter(ctx context.Context, chapterID uuid.UUID, params pagination.Params) (ListNotesResult, error) {
	return s.repository.ListAdminByChapter(ctx, chapterID, params)
}

func (s *Service) ListPublicByChapter(ctx context.Context, chapterID uuid.UUID, params pagination.Params) (ListPublicNotesResult, error) {
	result, err := s.repository.ListPublicByChapter(ctx, chapterID, params)
	if err != nil {
		return ListPublicNotesResult{}, err
	}

	publicNotes := make([]PublicNote, 0, len(result.Notes))

	for _, n := range result.Notes {
		publicNotes = append(publicNotes, s.toPublicNote(n))
	}

	return ListPublicNotesResult{
		Notes:      publicNotes,
		TotalItems: result.TotalItems,
	}, nil
}

func (s *Service) ListAdmin(ctx context.Context, params pagination.Params) (ListAdminNotesResult, error) {
	return s.repository.ListAdmin(ctx, params)
}

func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (Note, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetPublicByID(ctx context.Context, id uuid.UUID) (PublicNote, error) {
	n, err := s.repository.GetPublishedByID(ctx, id)
	if err != nil {
		return PublicNote{}, err
	}

	return s.toPublicNote(n), nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateNoteInput) (Note, error) {
	normalized := validation.NormalizeTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	})

	input.Title = normalized.Title
	input.Slug = normalized.Slug

	if err := validateTitleAndSlug(input.Title, input.Slug); err != nil {
		return Note{}, err
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	existingNote, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.objectStorage.DeleteObject(ctx, existingNote.StoredObjectKey); err != nil {
		return fmt.Errorf("delete note object from storage: %w", err)
	}

	return nil
}

func validateTitleAndSlug(title string, slug string) error {
	return validation.ValidateTitleSlug(validation.TitleSlugInput{
		Title: title,
		Slug:  slug,
	})
}

func detectContentType(file io.ReadSeeker) (string, error) {
	buffer := make([]byte, 512)

	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read file header: %w", err)
	}

	if n == 0 {
		return "", ErrInvalidFileSize
	}

	contentType := http.DetectContentType(buffer[:n])

	return contentType, nil
}

func isAllowedContentType(contentType string) bool {
	switch contentType {
	case "application/pdf",
		"image/jpeg",
		"image/png",
		"image/webp":
		return true
	default:
		return false
	}
}

func buildObjectKey(chapterID uuid.UUID, originalFilename string) string {
	extension := strings.ToLower(filepath.Ext(originalFilename))
	safeExtension := sanitizeExtension(extension)

	objectID := uuid.NewString()

	return filepath.Join(
		"notes",
		chapterID.String(),
		objectID+safeExtension,
	)
}

func sanitizeExtension(extension string) string {
	switch strings.ToLower(extension) {
	case ".pdf", ".jpg", ".jpeg", ".png", ".webp":
		return strings.ToLower(extension)
	default:
		return ""
	}
}

func ParseBoolFormValue(value string) bool {
	parsed, err := strconv.ParseBool(strings.TrimSpace(value))
	if err != nil {
		return false
	}

	return parsed
}

func (s *Service) GetPublishedFileMetadata(ctx context.Context, id uuid.UUID) (FileMetadata, error) {
	return s.repository.GetPublishedFileMetadata(ctx, id)
}

func (s *Service) toPublicNote(n Note) PublicNote {
	return PublicNote{
		ID:               n.ID,
		ChapterID:        n.ChapterID,
		Title:            n.Title,
		Slug:             n.Slug,
		Description:      n.Description,
		OriginalFileName: n.OriginalFileName,
		FileURL:          s.buildPublicFileURL(n.ID),
		IsPublished:      n.IsPublished,
		SortOrder:        n.SortOrder,
		CreatedAt:        n.CreatedAt,
		UpdatedAt:        n.UpdatedAt,
	}
}

func (s *Service) buildPublicFileURL(noteID uuid.UUID) string {
	if s.publicFileBaseURL == "" {
		return "/notes/" + noteID.String()
	}

	return s.publicFileBaseURL + "/notes/" + noteID.String()
}
