package semester

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrTitleRequired = errors.New("title is required")
	ErrSlugRequired  = errors.New("slug is required")
	ErrInvalidSlug   = errors.New("slug may only contain lowercase letters, numbers, and hyphens")
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

func (s *Service) Create(ctx context.Context, input CreateSemesterInput) (Semester, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)

	if err := validateTitleAndSlug(input.Title, input.Slug); err != nil {
		return Semester{}, err
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) ListAdmin(ctx context.Context) ([]Semester, error) {
	return s.repository.ListAdmin(ctx)
}

func (s *Service) ListPublic(ctx context.Context) ([]Semester, error) {
	return s.repository.ListPublic(ctx)
}

func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (Semester, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetPublicByID(ctx context.Context, id uuid.UUID) (Semester, error) {
	return s.repository.GetPublishedByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateSemesterInput) (Semester, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.TrimSpace(input.Slug)

	if err := validateTitleAndSlug(input.Title, input.Slug); err != nil {
		return Semester{}, err
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
