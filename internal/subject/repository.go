package subject

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSubjectNotFound      = errors.New("subject not found")
	ErrSubjectSlugConflicts = errors.New("subject slug already exists in this semester")
	ErrSemesterNotFound     = errors.New("semester not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, semesterID uuid.UUID, input CreateSubjectInput) (Subject, error) {
	const query = `
		INSERT INTO subjects (
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING
			id,
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_at,
			updated_at;
	`

	var subject Subject

	err := r.db.QueryRow(
		ctx,
		query,
		semesterID,
		input.Title,
		input.Slug,
		input.Description,
		input.SortOrder,
		input.IsPublished,
	).Scan(
		&subject.ID,
		&subject.SemesterID,
		&subject.Title,
		&subject.Slug,
		&subject.Description,
		&subject.SortOrder,
		&subject.IsPublished,
		&subject.CreatedAt,
		&subject.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Subject{}, ErrSubjectSlugConflicts
		}

		if isForeignKeyViolation(err) {
			return Subject{}, ErrSemesterNotFound
		}

		return Subject{}, fmt.Errorf("create subject: %w", err)
	}

	return subject, nil
}

func (r *Repository) ListAdminBySemester(ctx context.Context, semesterID uuid.UUID) ([]Subject, error) {
	const query = `
		SELECT
			id,
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_at,
			updated_at
		FROM subjects
		WHERE semester_id = $1
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query, semesterID)
}

func (r *Repository) ListPublicBySemester(ctx context.Context, semesterID uuid.UUID) ([]Subject, error) {
	const query = `
		SELECT
			id,
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_at,
			updated_at
		FROM subjects
		WHERE semester_id = $1
		AND is_published = true
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query, semesterID)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Subject, error) {
	const query = `
		SELECT
			id,
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_at,
			updated_at
		FROM subjects
		WHERE id = $1;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) GetPublishedByID(ctx context.Context, id uuid.UUID) (Subject, error) {
	const query = `
		SELECT
			id,
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_at,
			updated_at
		FROM subjects
		WHERE id = $1
		AND is_published = true;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateSubjectInput) (Subject, error) {
	const query = `
		UPDATE subjects
		SET
			title = $2,
			slug = $3,
			description = $4,
			sort_order = $5,
			is_published = $6,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id,
			semester_id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_at,
			updated_at;
	`

	var subject Subject

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Title,
		input.Slug,
		input.Description,
		input.SortOrder,
		input.IsPublished,
	).Scan(
		&subject.ID,
		&subject.SemesterID,
		&subject.Title,
		&subject.Slug,
		&subject.Description,
		&subject.SortOrder,
		&subject.IsPublished,
		&subject.CreatedAt,
		&subject.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subject{}, ErrSubjectNotFound
		}

		if isUniqueViolation(err) {
			return Subject{}, ErrSubjectSlugConflicts
		}

		return Subject{}, fmt.Errorf("update subject: %w", err)
	}

	return subject, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM subjects
		WHERE id = $1;
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subject: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSubjectNotFound
	}

	return nil
}

func (r *Repository) list(ctx context.Context, query string, args ...any) ([]Subject, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}
	defer rows.Close()

	subjects := make([]Subject, 0)

	for rows.Next() {
		var subject Subject

		if err := rows.Scan(
			&subject.ID,
			&subject.SemesterID,
			&subject.Title,
			&subject.Slug,
			&subject.Description,
			&subject.SortOrder,
			&subject.IsPublished,
			&subject.CreatedAt,
			&subject.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}

		subjects = append(subjects, subject)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subjects: %w", err)
	}

	return subjects, nil
}

func (r *Repository) getOne(ctx context.Context, query string, args ...any) (Subject, error) {
	var subject Subject

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&subject.ID,
		&subject.SemesterID,
		&subject.Title,
		&subject.Slug,
		&subject.Description,
		&subject.SortOrder,
		&subject.IsPublished,
		&subject.CreatedAt,
		&subject.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Subject{}, ErrSubjectNotFound
		}

		return Subject{}, fmt.Errorf("get subject: %w", err)
	}

	return subject, nil
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
