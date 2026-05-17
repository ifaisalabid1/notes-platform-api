package note

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ifaisalabid1/notes-platform-api/internal/pagination"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoteNotFound       = errors.New("note not found")
	ErrNoteSlugConflicts  = errors.New("note slug already exists in this chapter")
	ErrObjectKeyConflicts = errors.New("stored object key already exists")
	ErrChapterNotFound    = errors.New("chapter not found")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, chapterID uuid.UUID, input CreateNoteInput) (Note, error) {
	const query = `
		INSERT INTO notes (
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id,
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order,
			uploaded_by,
			created_at,
			updated_at;
	`

	var note Note

	err := r.db.QueryRow(
		ctx,
		query,
		chapterID,
		input.Title,
		input.Slug,
		input.Description,
		input.OriginalFileName,
		input.StoredObjectKey,
		input.FileContentType,
		input.FileSizeBytes,
		input.IsWatermarked,
		input.IsPublished,
		input.SortOrder,
	).Scan(
		&note.ID,
		&note.ChapterID,
		&note.Title,
		&note.Slug,
		&note.Description,
		&note.OriginalFileName,
		&note.StoredObjectKey,
		&note.FileContentType,
		&note.FileSizeBytes,
		&note.IsWatermarked,
		&note.IsPublished,
		&note.SortOrder,
		&note.UploadedBy,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err != nil {
		if isUniqueViolation(err) {
			return Note{}, classifyUniqueViolation(err)
		}

		if isForeignKeyViolation(err) {
			return Note{}, ErrChapterNotFound
		}

		return Note{}, fmt.Errorf("create note: %w", err)
	}

	return note, nil
}

func (r *Repository) ListAdminByChapter(ctx context.Context, chapterID uuid.UUID, params pagination.Params) (ListNotesResult, error) {
	const query = `
		SELECT
			id,
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order,
			uploaded_by,
			created_at,
			updated_at,
			COUNT(*) OVER()
		FROM notes
		WHERE chapter_id = $1
		AND (
			$2 = ''
			OR title ILIKE '%' || $2 || '%'
			OR slug ILIKE '%' || $2 || '%'
			OR original_file_name ILIKE '%' || $2 || '%'
		)
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $3 OFFSET $4;
	`

	return r.listPaginated(ctx, query, chapterID, params.Search, params.Limit(), params.Offset())
}

func (r *Repository) ListPublicByChapter(ctx context.Context, chapterID uuid.UUID, params pagination.Params) (ListNotesResult, error) {
	const query = `
		SELECT
			id,
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order,
			uploaded_by,
			created_at,
			updated_at,
			COUNT(*) OVER()
		FROM notes
		WHERE chapter_id = $1
		AND is_published = true
		AND (
			$2 = ''
			OR title ILIKE '%' || $2 || '%'
			OR slug ILIKE '%' || $2 || '%'
			OR original_file_name ILIKE '%' || $2 || '%'
		)
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $3 OFFSET $4;
	`

	return r.listPaginated(ctx, query, chapterID, params.Search, params.Limit(), params.Offset())
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Note, error) {
	const query = `
		SELECT
			id,
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order,
			uploaded_by,
			created_at,
			updated_at
		FROM notes
		WHERE id = $1;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) GetPublishedByID(ctx context.Context, id uuid.UUID) (Note, error) {
	const query = `
		SELECT
			id,
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order,
			uploaded_by,
			created_at,
			updated_at
		FROM notes
		WHERE id = $1
		AND is_published = true;
	`

	return r.getOne(ctx, query, id)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input UpdateNoteInput) (Note, error) {
	const query = `
		UPDATE notes
		SET
			title = $2,
			slug = $3,
			description = $4,
			is_published = $5,
			sort_order = $6,
			updated_at = now()
		WHERE id = $1
		RETURNING
			id,
			chapter_id,
			title,
			slug,
			description,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes,
			is_watermarked,
			is_published,
			sort_order,
			uploaded_by,
			created_at,
			updated_at;
	`

	var note Note

	err := r.db.QueryRow(
		ctx,
		query,
		id,
		input.Title,
		input.Slug,
		input.Description,
		input.IsPublished,
		input.SortOrder,
	).Scan(
		&note.ID,
		&note.ChapterID,
		&note.Title,
		&note.Slug,
		&note.Description,
		&note.OriginalFileName,
		&note.StoredObjectKey,
		&note.FileContentType,
		&note.FileSizeBytes,
		&note.IsWatermarked,
		&note.IsPublished,
		&note.SortOrder,
		&note.UploadedBy,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Note{}, ErrNoteNotFound
		}

		if isUniqueViolation(err) {
			return Note{}, classifyUniqueViolation(err)
		}

		return Note{}, fmt.Errorf("update note: %w", err)
	}

	return note, nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
		DELETE FROM notes
		WHERE id = $1;
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrNoteNotFound
	}

	return nil
}

func (r *Repository) list(ctx context.Context, query string, args ...any) ([]Note, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)

	for rows.Next() {
		var note Note

		if err := rows.Scan(
			&note.ID,
			&note.ChapterID,
			&note.Title,
			&note.Slug,
			&note.Description,
			&note.OriginalFileName,
			&note.StoredObjectKey,
			&note.FileContentType,
			&note.FileSizeBytes,
			&note.IsWatermarked,
			&note.IsPublished,
			&note.SortOrder,
			&note.UploadedBy,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return notes, nil
}

func (r *Repository) listPaginated(ctx context.Context, query string, args ...any) (ListNotesResult, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return ListNotesResult{}, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)
	totalItems := 0

	for rows.Next() {
		var note Note
		var rowTotal int

		if err := rows.Scan(
			&note.ID,
			&note.ChapterID,
			&note.Title,
			&note.Slug,
			&note.Description,
			&note.OriginalFileName,
			&note.StoredObjectKey,
			&note.FileContentType,
			&note.FileSizeBytes,
			&note.IsWatermarked,
			&note.IsPublished,
			&note.SortOrder,
			&note.UploadedBy,
			&note.CreatedAt,
			&note.UpdatedAt,
			&rowTotal,
		); err != nil {
			return ListNotesResult{}, fmt.Errorf("scan note: %w", err)
		}

		totalItems = rowTotal
		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		return ListNotesResult{}, fmt.Errorf("iterate notes: %w", err)
	}

	return ListNotesResult{
		Notes:      notes,
		TotalItems: totalItems,
	}, nil
}

func (r *Repository) getOne(ctx context.Context, query string, args ...any) (Note, error) {
	var note Note

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&note.ID,
		&note.ChapterID,
		&note.Title,
		&note.Slug,
		&note.Description,
		&note.OriginalFileName,
		&note.StoredObjectKey,
		&note.FileContentType,
		&note.FileSizeBytes,
		&note.IsWatermarked,
		&note.IsPublished,
		&note.SortOrder,
		&note.UploadedBy,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Note{}, ErrNoteNotFound
		}

		return Note{}, fmt.Errorf("get note: %w", err)
	}

	return note, nil
}

func classifyUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "notes_chapter_id_slug_key":
			return ErrNoteSlugConflicts
		case "notes_stored_object_key_key":
			return ErrObjectKeyConflicts
		}
	}

	return ErrNoteSlugConflicts
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

func (r *Repository) GetPublishedFileMetadata(ctx context.Context, id uuid.UUID) (FileMetadata, error) {
	const query = `
		SELECT
			id,
			title,
			original_file_name,
			stored_object_key,
			file_content_type,
			file_size_bytes
		FROM notes
		WHERE id = $1
		AND is_published = true;
	`

	var metadata FileMetadata

	err := r.db.QueryRow(ctx, query, id).Scan(
		&metadata.NoteID,
		&metadata.Title,
		&metadata.OriginalFileName,
		&metadata.StoredObjectKey,
		&metadata.FileContentType,
		&metadata.FileSizeBytes,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FileMetadata{}, ErrNoteNotFound
		}

		return FileMetadata{}, fmt.Errorf("get published file metadata: %w", err)
	}

	return metadata, nil
}
