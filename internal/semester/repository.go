package semester

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
	ErrSemesterNotFound      = errors.New("semester not found")
	ErrSemesterSlugConflicts = errors.New("semester slug already exists")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, input CreateSemesterInput) (Semester, error) {
	const query = `
		INSERT INTO semesters (
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING
			id,
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
	var semester Semester

	err := r.db.QueryRow(
		ctx,
		query,
		input.Title,
		input.Slug,
		input.Description,
		input.SortOrder,
		input.IsPublished,
		actorID,
	).Scan(
		&semester.ID,
		&semester.Title,
		&semester.Slug,
		&semester.Description,
		&semester.SortOrder,
		&semester.IsPublished,
		&semester.CreatedBy,
		&semester.UpdatedBy,
		&semester.CreatedAt,
		&semester.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Semester{}, ErrSemesterSlugConflicts
		}

		return Semester{}, fmt.Errorf("create semester: %w", err)
	}

	return semester, nil
}

func (r *Repository) ListAdmin(ctx context.Context) ([]Semester, error) {
	const query = `
		SELECT
			id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM semesters
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query)
}

func (r *Repository) ListPublic(ctx context.Context) ([]Semester, error) {
	const query = `
		SELECT
			id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM semesters
		WHERE is_published = true
		ORDER BY sort_order ASC, created_at DESC;
	`

	return r.list(ctx, query)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Semester, error) {
	const query = `
		SELECT
			id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM semesters
		WHERE id = $1;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) GetPublishedByID(ctx context.Context, id uuid.UUID) (Semester, error) {
	const query = `
		SELECT
			id,
			title,
			slug,
			description,
			sort_order,
			is_published,
			created_by,
			updated_by,
			created_at,
			updated_at
		FROM semesters
		WHERE id = $1
		AND is_published = true;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateSemesterInput) (Semester, error) {
	const query = `
		UPDATE semesters
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

	var semester Semester

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
		&semester.ID,
		&semester.Title,
		&semester.Slug,
		&semester.Description,
		&semester.SortOrder,
		&semester.IsPublished,
		&semester.CreatedBy,
		&semester.UpdatedBy,
		&semester.CreatedAt,
		&semester.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Semester{}, ErrSemesterNotFound
		}

		if isUniqueViolation(err) {
			return Semester{}, ErrSemesterSlugConflicts
		}

		return Semester{}, fmt.Errorf("update semester: %w", err)
	}

	return semester, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM semesters
		WHERE id = $1;
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete semester: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSemesterNotFound
	}

	return nil
}

func (r *Repository) list(ctx context.Context, query string) ([]Semester, error) {
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list semesters: %w", err)
	}
	defer rows.Close()

	semesters := make([]Semester, 0)

	for rows.Next() {
		var semester Semester

		if err := rows.Scan(
			&semester.ID,
			&semester.Title,
			&semester.Slug,
			&semester.Description,
			&semester.SortOrder,
			&semester.IsPublished,
			&semester.CreatedBy,
			&semester.UpdatedBy,
			&semester.CreatedAt,
			&semester.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan semester: %w", err)
		}

		semesters = append(semesters, semester)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate semesters: %w", err)
	}

	return semesters, nil
}

func (r *Repository) getOne(ctx context.Context, query string, args ...any) (Semester, error) {
	var semester Semester

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&semester.ID,
		&semester.Title,
		&semester.Slug,
		&semester.Description,
		&semester.SortOrder,
		&semester.IsPublished,
		&semester.CreatedBy,
		&semester.UpdatedBy,
		&semester.CreatedAt,
		&semester.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Semester{}, ErrSemesterNotFound
		}

		return Semester{}, fmt.Errorf("get semester: %w", err)
	}

	return semester, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}

func (r *Repository) ListAdminWithAudit(ctx context.Context) ([]AdminSemester, error) {
	const query = `
		SELECT
			s.id,
			s.title,
			s.slug,
			s.description,
			s.sort_order,
			s.is_published,
			s.created_by,
			created_admin.display_name AS created_by_name,
			s.updated_by,
			updated_admin.display_name AS updated_by_name,
			s.created_at,
			s.updated_at
		FROM semesters s
		LEFT JOIN admins created_admin ON created_admin.id = s.created_by
		LEFT JOIN admins updated_admin ON updated_admin.id = s.updated_by
		ORDER BY s.sort_order ASC, s.created_at DESC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list admin semesters with audit: %w", err)
	}
	defer rows.Close()

	semesters := make([]AdminSemester, 0)

	for rows.Next() {
		var semester AdminSemester

		if err := rows.Scan(
			&semester.ID,
			&semester.Title,
			&semester.Slug,
			&semester.Description,
			&semester.SortOrder,
			&semester.IsPublished,
			&semester.CreatedBy,
			&semester.CreatedByName,
			&semester.UpdatedBy,
			&semester.UpdatedByName,
			&semester.CreatedAt,
			&semester.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan admin semester: %w", err)
		}

		semesters = append(semesters, semester)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admin semesters: %w", err)
	}

	return semesters, nil
}

func (r *Repository) GetAdminWithAuditByID(ctx context.Context, id uuid.UUID) (AdminSemester, error) {
	const query = `
		SELECT
			s.id,
			s.title,
			s.slug,
			s.description,
			s.sort_order,
			s.is_published,
			s.created_by,
			created_admin.display_name AS created_by_name,
			s.updated_by,
			updated_admin.display_name AS updated_by_name,
			s.created_at,
			s.updated_at
		FROM semesters s
		LEFT JOIN admins created_admin ON created_admin.id = s.created_by
		LEFT JOIN admins updated_admin ON updated_admin.id = s.updated_by
		WHERE s.id = $1;
	`

	var semester AdminSemester

	err := r.db.QueryRow(ctx, query, id).Scan(
		&semester.ID,
		&semester.Title,
		&semester.Slug,
		&semester.Description,
		&semester.SortOrder,
		&semester.IsPublished,
		&semester.CreatedBy,
		&semester.CreatedByName,
		&semester.UpdatedBy,
		&semester.UpdatedByName,
		&semester.CreatedAt,
		&semester.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminSemester{}, ErrSemesterNotFound
		}

		return AdminSemester{}, fmt.Errorf("get admin semester with audit: %w", err)
	}

	return semester, nil
}
