package note

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/storage"
)

var (
	ErrTitleRequired            = errors.New("title is required")
	ErrSlugRequired             = errors.New("slug is required")
	ErrInvalidSlug              = errors.New("slug may only contain lowercase letters, numbers, and hyphens")
	ErrOriginalFileNameRequired = errors.New("original file name is required")
	ErrStoredObjectKeyRequired  = errors.New("stored object key is required")
	ErrFileContentTypeRequired  = errors.New("file content type is required")
	ErrInvalidFileSize          = errors.New("file size must be greater than zero")
	ErrFileRequired             = errors.New("file is required")
	ErrUnsupportedFileType      = errors.New("unsupported file type")
	ErrFileTooLarge             = errors.New("file is too large")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	repository     *Repository
	objectStorage  storage.ObjectStorage
	uploadMaxBytes int64
}

func NewService(repository *Repository, objectStorage storage.ObjectStorage, uploadMaxBytes int64) *Service {
	return &Service{
		repository:     repository,
		objectStorage:  objectStorage,
		uploadMaxBytes: uploadMaxBytes,
	}
}

func (s *Service) Create(ctx context.Context, chapterID uuid.UUID, input CreateNoteInput) (Note, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)
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
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)

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

	objectKey := buildObjectKey(chapterID, input.FileHeader.Filename)

	putResult, err := s.objectStorage.PutObject(ctx, storage.PutObjectInput{
		Key:         objectKey,
		Body:        input.File,
		ContentType: contentType,
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
		FileContentType:  contentType,
		FileSizeBytes:    putResult.SizeBytes,
		IsWatermarked:    false,
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

func (s *Service) ListAdminByChapter(ctx context.Context, chapterID uuid.UUID) ([]Note, error) {
	return s.repository.ListAdminByChapter(ctx, chapterID)
}

func (s *Service) ListPublicByChapter(ctx context.Context, chapterID uuid.UUID) ([]Note, error) {
	return s.repository.ListPublicByChapter(ctx, chapterID)
}

func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (Note, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetPublicByID(ctx context.Context, id uuid.UUID) (Note, error) {
	return s.repository.GetPublishedByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateNoteInput) (Note, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)

	if err := validateTitleAndSlug(input.Title, input.Slug); err != nil {
		return Note{}, err
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}

func validateTitleAndSlug(title string, slug string) error {
	if title == "" {
		return ErrTitleRequired
	}

	if slug == "" {
		return ErrSlugRequired
	}

	if !slugPattern.MatchString(slug) {
		return ErrInvalidSlug
	}

	return nil
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
