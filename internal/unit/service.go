package unit

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

func (s *Service) Create(ctx context.Context, subjectID uuid.UUID, input CreateUnitInput) (Unit, error) {
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
		return Unit{}, err
	}

	return s.repository.Create(ctx, subjectID, input)
}

func (s *Service) ListAdminBySubject(ctx context.Context, subjectID uuid.UUID) ([]AdminUnit, error) {
	return s.repository.ListAdminBySubjectWithAudit(ctx, subjectID)
}

func (s *Service) ListPublicBySubject(ctx context.Context, subjectID uuid.UUID) ([]Unit, error) {
	return s.repository.ListPublicBySubject(ctx, subjectID)
}

func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (AdminUnit, error) {
	return s.repository.GetAdminWithAuditByID(ctx, id)
}

func (s *Service) GetPublicByID(ctx context.Context, id uuid.UUID) (Unit, error) {
	return s.repository.GetPublishedByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateUnitInput) (Unit, error) {
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
		return Unit{}, err
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
