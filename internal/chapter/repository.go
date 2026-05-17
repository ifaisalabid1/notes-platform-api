package chapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/notes-platform-api/internal/audit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrChapterNotFound      = errors.New("chapter not found")
	ErrChapterSlugConflicts = errors.New("chapter slug already exists in this unit")
	ErrUnitNotFound         = errors.New("unit not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, unitID uuid.UUID, input CreateChapterInput) (Chapter, error) {
	const query = `
		INSERT INTO chapters (
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING
			id,
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at;
	`

	actorID := audit.ActorIDFromContext(ctx)
	var chapter Chapter

	err := r.db.QueryRow(
		ctx,
		query,
		unitID,
		input.Title,
		input.Slug,
		input.Description,
		input.SortOrder,
		input.IsPublished,
		actorID,
	).Scan(
		&chapter.ID,
		&chapter.UnitID,
		&chapter.Title,
		&chapter.Slug,
		&chapter.Description,
		&chapter.SortOrder,
		&chapter.IsPublished,
		&chapter.CreatedBy,
		&chapter.UpdatedBy,
		&chapter.CreatedAt,
		&chapter.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Chapter{}, ErrChapterSlugConflicts
		}

		if isForeignKeyViolation(err) {
			return Chapter{}, ErrUnitNotFound
		}

		return Chapter{}, fmt.Errorf("create chapter: %w", err)
	}

	return chapter, nil
}

func (r *Repository) ListAdminByUnit(ctx context.Context, unitID uuid.UUID) ([]Chapter, error) {
	const query = `
		SELECT
			id,
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM chapters
		WHERE unit_id = $1
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query, unitID)
}

func (r *Repository) ListPublicByUnit(ctx context.Context, unitID uuid.UUID) ([]Chapter, error) {
	const query = `
		SELECT
			id,
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM chapters
		WHERE unit_id = $1
		AND is_published = true
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query, unitID)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Chapter, error) {
	const query = `
		SELECT
			id,
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM chapters
		WHERE id = $1;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) GetPublishedByID(ctx context.Context, id uuid.UUID) (Chapter, error) {
	const query = `
		SELECT
			id,
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM chapters
		WHERE id = $1
		AND is_published = true;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateChapterInput) (Chapter, error) {
	const query = `
		UPDATE chapters
		SET
			title = $2,
			slug = $3,
			description = $4,
			sort_order = $5,
			is_published = $6,
			updated_by = $7,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id,
			unit_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at;
	`

	actorID := audit.ActorIDFromContext(ctx)
	var chapter Chapter

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Title,
		input.Slug,
		input.Description,
		input.SortOrder,
		input.IsPublished,
		actorID,
	).Scan(
		&chapter.ID,
		&chapter.UnitID,
		&chapter.Title,
		&chapter.Slug,
		&chapter.Description,
		&chapter.SortOrder,
		&chapter.IsPublished,
		&chapter.CreatedBy,
		&chapter.UpdatedBy,
		&chapter.CreatedAt,
		&chapter.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Chapter{}, ErrChapterNotFound
		}

		if isUniqueViolation(err) {
			return Chapter{}, ErrChapterSlugConflicts
		}

		return Chapter{}, fmt.Errorf("update chapter: %w", err)
	}

	return chapter, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM chapters
		WHERE id = $1;
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete chapter: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrChapterNotFound
	}

	return nil
}

func (r *Repository) list(ctx context.Context, query string, args ...any) ([]Chapter, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	chapters := make([]Chapter, 0)

	for rows.Next() {
		var chapter Chapter

		if err := rows.Scan(
			&chapter.ID,
			&chapter.UnitID,
			&chapter.Title,
			&chapter.Slug,
			&chapter.Description,
			&chapter.SortOrder,
			&chapter.IsPublished,
			&chapter.CreatedBy,
			&chapter.UpdatedBy,
			&chapter.CreatedAt,
			&chapter.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}

		chapters = append(chapters, chapter)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chapters: %w", err)
	}

	return chapters, nil
}

func (r *Repository) getOne(ctx context.Context, query string, args ...any) (Chapter, error) {
	var chapter Chapter

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&chapter.ID,
		&chapter.UnitID,
		&chapter.Title,
		&chapter.Slug,
		&chapter.Description,
		&chapter.SortOrder,
		&chapter.IsPublished,
		&chapter.CreatedBy,
		&chapter.UpdatedBy,
		&chapter.CreatedAt,
		&chapter.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Chapter{}, ErrChapterNotFound
		}

		return Chapter{}, fmt.Errorf("get chapter: %w", err)
	}

	return chapter, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}

	return false
}

func (r *Repository) ListAdminByUnitWithAudit(ctx context.Context, unitID uuid.UUID) ([]AdminChapter, error) {
	const query = `
		SELECT
			c.id,
			c.unit_id,
			c.title,
			c.slug,
			c.description,
			c.sort_order,
			c.is_published,
			c.created_by,
			created_admin.display_name AS created_by_name,
			c.updated_by,
			updated_admin.display_name AS updated_by_name,
			c.created_at,
			c.updated_at
		FROM chapters c
		LEFT JOIN admins created_admin ON created_admin.id = c.created_by
		LEFT JOIN admins updated_admin ON updated_admin.id = c.updated_by
		WHERE c.unit_id = $1
		ORDER BY c.sort_order ASC, c.created_at DESC;
	`

	rows, err := r.db.Query(ctx, query, unitID)
	if err != nil {
		return nil, fmt.Errorf("list admin chapters with audit: %w", err)
	}
	defer rows.Close()

	chapters := make([]AdminChapter, 0)

	for rows.Next() {
		var chapter AdminChapter

		if err := rows.Scan(
			&chapter.ID,
			&chapter.UnitID,
			&chapter.Title,
			&chapter.Slug,
			&chapter.Description,
			&chapter.SortOrder,
			&chapter.IsPublished,
			&chapter.CreatedBy,
			&chapter.CreatedByName,
			&chapter.UpdatedBy,
			&chapter.UpdatedByName,
			&chapter.CreatedAt,
			&chapter.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan admin chapter: %w", err)
		}

		chapters = append(chapters, chapter)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin chapters: %w", err)
	}

	return chapters, nil
}

func (r *Repository) GetAdminWithAuditByID(ctx context.Context, id uuid.UUID) (AdminChapter, error) {
	const query = `
		SELECT
			c.id,
			c.unit_id,
			c.title,
			c.slug,
			c.description,
			c.sort_order,
			c.is_published,
			c.created_by,
			created_admin.display_name AS created_by_name,
			c.updated_by,
			updated_admin.display_name AS updated_by_name,
			c.created_at,
			c.updated_at
		FROM chapters c
		LEFT JOIN admins created_admin ON created_admin.id = c.created_by
		LEFT JOIN admins updated_admin ON updated_admin.id = c.updated_by
		WHERE c.id = $1;
	`

	var chapter AdminChapter

	err := r.db.QueryRow(ctx, query, id).Scan(
		&chapter.ID,
		&chapter.UnitID,
		&chapter.Title,
		&chapter.Slug,
		&chapter.Description,
		&chapter.SortOrder,
		&chapter.IsPublished,
		&chapter.CreatedBy,
		&chapter.CreatedByName,
		&chapter.UpdatedBy,
		&chapter.UpdatedByName,
		&chapter.CreatedAt,
		&chapter.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminChapter{}, ErrChapterNotFound
		}

		return AdminChapter{}, fmt.Errorf("get admin chapter with audit: %w", err)
	}

	return chapter, nil
}
