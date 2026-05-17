package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAdminNotFound   = errors.New("admin not found")
	ErrEmailConflict   = errors.New("admin email already exists")
	ErrOwnerExists     = errors.New("owner admin already exists")
	ErrOwnerNotAllowed = errors.New("email is not allowed to become owner")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CountOwners(ctx context.Context) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM admins
		WHERE role = 'owner';
	`

	var count int

	if err := r.db.QueryRow(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("count owners: %w", err)
	}

	return count, nil
}

func (r *Repository) CreateOwner(ctx context.Context, email string, passwordHash string, displayName string) (Admin, error) {
	const query = `
		INSERT INTO admins (
			email,
			password_hash,
			display_name,
			role
		)
		VALUES ($1, $2, $3, 'owner')
		RETURNING
			id,
			email,
			display_name,
			role,
			is_active,
			last_login_at,
			created_at,
			updated_at;
	`

	return r.create(ctx, query, email, passwordHash, displayName)
}

func (r *Repository) CreateAdmin(ctx context.Context, email string, passwordHash string, displayName string, createdBy uuid.UUID) (Admin, error) {
	const query = `
		INSERT INTO admins (
			email,
			password_hash,
			display_name,
			role,
			created_by
		)
		VALUES ($1, $2, $3, 'admin', $4)
		RETURNING
			id,
			email,
			display_name,
			role,
			is_active,
			last_login_at,
			created_at,
			updated_at;
	`

	var a Admin

	err := r.db.QueryRow(
		ctx,
		query,
		email,
		passwordHash,
		displayName,
		createdBy,
	).Scan(
		&a.ID,
		&a.Email,
		&a.DisplayName,
		&a.Role,
		&a.IsActive,
		&a.LastLoginAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Admin{}, ErrEmailConflict
		}

		return Admin{}, fmt.Errorf("create admin: %w", err)
	}

	return a, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (AdminWithPassword, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash,
			display_name,
			role,
			is_active,
			last_login_at,
			created_at,
			updated_at
		FROM admins
		WHERE email = $1;
	`

	var a AdminWithPassword

	err := r.db.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&a.ID,
		&a.Email,
		&a.PasswordHash,
		&a.DisplayName,
		&a.Role,
		&a.IsActive,
		&a.LastLoginAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdminWithPassword{}, ErrAdminNotFound
		}

		return AdminWithPassword{}, fmt.Errorf("get admin by email: %w", err)
	}

	return a, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Admin, error) {
	const query = `
		SELECT
			id,
			email,
			display_name,
			role,
			is_active,
			last_login_at,
			created_at,
			updated_at
		FROM admins
		WHERE id = $1;
	`

	var a Admin

	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID,
		&a.Email,
		&a.DisplayName,
		&a.Role,
		&a.IsActive,
		&a.LastLoginAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Admin{}, ErrAdminNotFound
		}

		return Admin{}, fmt.Errorf("get admin by id: %w", err)
	}

	return a, nil
}

func (r *Repository) TouchLastLogin(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE admins
		SET
			last_login_at = now(),
			updated_at = now()
		WHERE id = $1;
	`

	if _, err := r.db.Exec(ctx, query, id); err != nil {
		return fmt.Errorf("touch last login: %w", err)
	}

	return nil
}

func (r *Repository) create(ctx context.Context, query string, email string, passwordHash string, displayName string) (Admin, error) {
	var a Admin

	err := r.db.QueryRow(
		ctx,
		query,
		strings.ToLower(email),
		passwordHash,
		displayName,
	).Scan(
		&a.ID,
		&a.Email,
		&a.DisplayName,
		&a.Role,
		&a.IsActive,
		&a.LastLoginAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Admin{}, ErrEmailConflict
		}

		return Admin{}, fmt.Errorf("create owner: %w", err)
	}

	return a, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}

func (r *Repository) List(ctx context.Context) ([]Admin, error) {
	const query = `
		SELECT
			id,
			email,
			display_name,
			role,
			is_active,
			last_login_at,
			created_at,
			updated_at
		FROM admins
		ORDER BY
			CASE role
				WHEN 'owner' THEN 0
				ELSE 1
			END,
			created_at ASC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list admins: %w", err)
	}
	defer rows.Close()

	admins := make([]Admin, 0)

	for rows.Next() {
		var a Admin

		if err := rows.Scan(
			&a.ID,
			&a.Email,
			&a.DisplayName,
			&a.Role,
			&a.IsActive,
			&a.LastLoginAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan admin: %w", err)
		}

		admins = append(admins, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate admins: %w", err)
	}

	return admins, nil
}
