package note

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired            = errors.New("title is required")
	ErrSlugRequired             = errors.New("slug is required")
	ErrInvalidSlug              = errors.New("slug may only contain lowercase letters, numbers, and hyphens")
	ErrOriginalFileNameRequired = errors.New("original file name is required")
	ErrStoredObjectKeyRequired  = errors.New("stored object key is required")
	ErrFileContentTypeRequired  = errors.New("file content type is required")
	ErrInvalidFileSize          = errors.New("file size must be greater than zero")
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
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
