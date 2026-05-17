package subject

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

func (s *Service) Create(ctx context.Context, semesterID uuid.UUID, input CreateSubjectInput) (Subject, error) {
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
		return Subject{}, err
	}

	return s.repository.Create(ctx, semesterID, input)
}

func (s *Service) ListAdminBySemester(ctx context.Context, semesterID uuid.UUID) ([]Subject, error) {
	return s.repository.ListAdminBySemester(ctx, semesterID)
}

func (s *Service) ListPublicBySemester(ctx context.Context, semesterID uuid.UUID) ([]Subject, error) {
	return s.repository.ListPublicBySemester(ctx, semesterID)
}

func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (Subject, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) GetPublicByID(ctx context.Context, id uuid.UUID) (Subject, error) {
	return s.repository.GetPublishedByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateSubjectInput) (Subject, error) {
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
		return Subject{}, err
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
