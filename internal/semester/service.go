package semester

import (
	"context"
	"errors"
	"regexp"
	"strings"
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

	if input.Title == "" {
		return Semester{}, ErrTitleRequired
	}

	if input.Slug == "" {
		return Semester{}, ErrSlugRequired
	}

	if !slugPattern.MatchString(input.Slug) {
		return Semester{}, ErrInvalidSlug
	}

	return s.repository.Create(ctx, input)
}

func (s *Service) ListAdmin(ctx context.Context) ([]Semester, error) {
	return s.repository.ListAdmin(ctx)
}

func (s *Service) ListPublic(ctx context.Context) ([]Semester, error) {
	return s.repository.ListPublic(ctx)
}
