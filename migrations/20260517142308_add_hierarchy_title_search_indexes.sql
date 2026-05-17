-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_chapters_title_trgm
ON chapters USING gin (title gin_trgm_ops);

CREATE INDEX idx_units_title_trgm
ON units USING gin (title gin_trgm_ops);

CREATE INDEX idx_subjects_title_trgm
ON subjects USING gin (title gin_trgm_ops);

CREATE INDEX idx_semesters_title_trgm
ON semesters USING gin (title gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_semesters_title_trgm;
DROP INDEX IF EXISTS idx_subjects_title_trgm;
DROP INDEX IF EXISTS idx_units_title_trgm;
DROP INDEX IF EXISTS idx_chapters_title_trgm;

-- +goose StatementEnd