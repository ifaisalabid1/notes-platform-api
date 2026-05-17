-- +goose Up
-- +goose StatementBegin

CREATE INDEX idx_notes_created_at
ON notes (created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_notes_created_at;

-- +goose StatementEnd