-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_admins_display_name_trgm
ON admins USING gin (display_name gin_trgm_ops);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_admins_display_name_trgm;

-- +goose StatementEnd