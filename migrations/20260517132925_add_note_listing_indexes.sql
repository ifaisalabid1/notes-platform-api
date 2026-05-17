-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_notes_chapter_sort_created
ON notes (chapter_id, sort_order ASC, created_at DESC);

CREATE INDEX idx_notes_published_chapter_sort_created
ON notes (chapter_id, sort_order ASC, created_at DESC)
WHERE is_published = true;

CREATE INDEX idx_notes_title_trgm
ON notes USING gin (title gin_trgm_ops);

CREATE INDEX idx_notes_slug_trgm
ON notes USING gin (slug gin_trgm_ops);

CREATE INDEX idx_notes_original_file_name_trgm
ON notes USING gin (original_file_name gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_notes_original_file_name_trgm;
DROP INDEX IF EXISTS idx_notes_slug_trgm;
DROP INDEX IF EXISTS idx_notes_title_trgm;
DROP INDEX IF EXISTS idx_notes_published_chapter_sort_created;
DROP INDEX IF EXISTS idx_notes_chapter_sort_created;

DROP EXTENSION IF EXISTS pg_trgm;

-- +goose StatementEnd