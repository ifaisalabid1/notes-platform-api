package chapter

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

func (s *Service) Create(ctx context.Context, unitID uuid.UUID, input CreateChapterInput) (Chapter, error) {
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
		return Chapter{}, err
	}

	return s.repository.Create(ctx, unitID, input)
}

func (s *Service) ListAdminByUnit(ctx context.Context, unitID uuid.UUID) ([]AdminChapter, error) {
	return s.repository.ListAdminByUnitWithAudit(ctx, unitID)
}

func (s *Service) ListPublicByUnit(ctx context.Context, unitID uuid.UUID) ([]Chapter, error) {
	return s.repository.ListPublicByUnit(ctx, unitID)
}

func (s *Service) GetAdminByID(ctx context.Context, id uuid.UUID) (AdminChapter, error) {
	return s.repository.GetAdminWithAuditByID(ctx, id)
}

func (s *Service) GetPublicByID(ctx context.Context, id uuid.UUID) (Chapter, error) {
	return s.repository.GetPublishedByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateChapterInput) (Chapter, error) {
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
		return Chapter{}, err
	}

	return s.repository.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}
