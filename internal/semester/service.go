package semester

import (
	"context"

	"github.com/google/uuid"

	"github.com/ifaisalabid1/notes-platform-api/internal/validation"
)

var (
	ErrTitleRequired = validation.ErrTitleRequired
	ErrSlugRequired  = validation.ErrSlugRequired
	ErrInvalidSlug   = validation.ErrInvalidSlug
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) Create(ctx context.Context, input CreateSemesterInput) (Semester, error) {
	normalized := validation.NormalizeTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	})

	input.Title = normalized.Title
	input.Slug = normalized.Slug

	if err := validation.ValidateTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	}); err != nil {
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
	normalized := validation.NormalizeTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	})

	input.Title = normalized.Title
	input.Slug = normalized.Slug

	if err := validation.ValidateTitleSlug(validation.TitleSlugInput{
		Title: input.Title,
		Slug:  input.Slug,
	}); err != nil {
		return Semester{}, err
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
