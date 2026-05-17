package unit

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
	ErrUnitNotFound      = errors.New("unit not found")
	ErrUnitSlugConflicts = errors.New("unit slug already exists in this subject")
	ErrSubjectNotFound   = errors.New("subject not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, subjectID uuid.UUID, input CreateUnitInput) (Unit, error) {
	const query = `
		INSERT INTO units (
			subject_id,
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
			subject_id,
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
	var unit Unit

	err := r.db.QueryRow(
		ctx,
		query,
		subjectID,
		input.Title,
		input.Slug,
		input.Description,
		input.SortOrder,
		input.IsPublished,
		actorID,
	).Scan(
		&unit.ID,
		&unit.SubjectID,
		&unit.Title,
		&unit.Slug,
		&unit.Description,
		&unit.SortOrder,
		&unit.IsPublished,
		&unit.CreatedBy,
		&unit.UpdatedBy,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Unit{}, ErrUnitSlugConflicts
		}

		if isForeignKeyViolation(err) {
			return Unit{}, ErrSubjectNotFound
		}

		return Unit{}, fmt.Errorf("create unit: %w", err)
	}

	return unit, nil
}

func (r *Repository) ListAdminBySubject(ctx context.Context, subjectID uuid.UUID) ([]Unit, error) {
	const query = `
		SELECT
			id,
			subject_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM units
		WHERE subject_id = $1
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query, subjectID)
}

func (r *Repository) ListPublicBySubject(ctx context.Context, subjectID uuid.UUID) ([]Unit, error) {
	const query = `
		SELECT
			id,
			subject_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM units
		WHERE subject_id = $1
		AND is_published = true
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query, subjectID)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Unit, error) {
	const query = `
		SELECT
			id,
			subject_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM units
		WHERE id = $1;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) GetPublishedByID(ctx context.Context, id uuid.UUID) (Unit, error) {
	const query = `
		SELECT
			id,
			subject_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM units
		WHERE id = $1
		AND is_published = true;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateUnitInput) (Unit, error) {
	const query = `
		UPDATE units
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
			subject_id,
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
	var unit Unit

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
		&unit.ID,
		&unit.SubjectID,
		&unit.Title,
		&unit.Slug,
		&unit.Description,
		&unit.SortOrder,
		&unit.IsPublished,
		&unit.CreatedBy,
		&unit.UpdatedBy,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Unit{}, ErrUnitNotFound
		}

		if isUniqueViolation(err) {
			return Unit{}, ErrUnitSlugConflicts
		}

		return Unit{}, fmt.Errorf("update unit: %w", err)
	}

	return unit, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM units
		WHERE id = $1;
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete unit: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUnitNotFound
	}

	return nil
}

func (r *Repository) list(ctx context.Context, query string, args ...any) ([]Unit, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()

	units := make([]Unit, 0)

	for rows.Next() {
		var unit Unit

		if err := rows.Scan(
			&unit.ID,
			&unit.SubjectID,
			&unit.Title,
			&unit.Slug,
			&unit.Description,
			&unit.SortOrder,
			&unit.IsPublished,
			&unit.CreatedBy,
			&unit.UpdatedBy,
			&unit.CreatedAt,
			&unit.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan unit: %w", err)
		}

		units = append(units, unit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate units: %w", err)
	}

	return units, nil
}

func (r *Repository) getOne(ctx context.Context, query string, args ...any) (Unit, error) {
	var unit Unit

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&unit.ID,
		&unit.SubjectID,
		&unit.Title,
		&unit.Slug,
		&unit.Description,
		&unit.SortOrder,
		&unit.IsPublished,
		&unit.CreatedBy,
		&unit.UpdatedBy,
		&unit.CreatedAt,
		&unit.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Unit{}, ErrUnitNotFound
		}

		return Unit{}, fmt.Errorf("get unit: %w", err)
	}

	return unit, nil
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
