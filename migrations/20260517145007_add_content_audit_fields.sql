-- +goose Up
-- +goose StatementBegin

ALTER TABLE semesters
ADD COLUMN created_by UUID REFERENCES admins(id) ON DELETE SET NULL,
ADD COLUMN updated_by UUID REFERENCES admins(id) ON DELETE SET NULL;

ALTER TABLE subjects
ADD COLUMN created_by UUID REFERENCES admins(id) ON DELETE SET NULL,
ADD COLUMN updated_by UUID REFERENCES admins(id) ON DELETE SET NULL;

ALTER TABLE units
ADD COLUMN created_by UUID REFERENCES admins(id) ON DELETE SET NULL,
ADD COLUMN updated_by UUID REFERENCES admins(id) ON DELETE SET NULL;

ALTER TABLE chapters
ADD COLUMN created_by UUID REFERENCES admins(id) ON DELETE SET NULL,
ADD COLUMN updated_by UUID REFERENCES admins(id) ON DELETE SET NULL;

ALTER TABLE notes
ADD COLUMN updated_by UUID REFERENCES admins(id) ON DELETE SET NULL;

CREATE INDEX idx_semesters_created_by ON semesters(created_by);
CREATE INDEX idx_semesters_updated_by ON semesters(updated_by);

CREATE INDEX idx_subjects_created_by ON subjects(created_by);
CREATE INDEX idx_subjects_updated_by ON subjects(updated_by);

CREATE INDEX idx_units_created_by ON units(created_by);
CREATE INDEX idx_units_updated_by ON units(updated_by);

CREATE INDEX idx_chapters_created_by ON chapters(created_by);
CREATE INDEX idx_chapters_updated_by ON chapters(updated_by);

CREATE INDEX idx_notes_uploaded_by ON notes(uploaded_by);
CREATE INDEX idx_notes_updated_by ON notes(updated_by);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_notes_updated_by;
DROP INDEX IF EXISTS idx_notes_uploaded_by;

DROP INDEX IF EXISTS idx_chapters_updated_by;
DROP INDEX IF EXISTS idx_chapters_created_by;

DROP INDEX IF EXISTS idx_units_updated_by;
DROP INDEX IF EXISTS idx_units_created_by;

DROP INDEX IF EXISTS idx_subjects_updated_by;
DROP INDEX IF EXISTS idx_subjects_created_by;

DROP INDEX IF EXISTS idx_semesters_updated_by;
DROP INDEX IF EXISTS idx_semesters_created_by;

ALTER TABLE notes
DROP COLUMN IF EXISTS updated_by;

ALTER TABLE chapters
DROP COLUMN IF EXISTS updated_by,
DROP COLUMN IF EXISTS created_by;

ALTER TABLE units
DROP COLUMN IF EXISTS updated_by,
DROP COLUMN IF EXISTS created_by;

ALTER TABLE subjects
DROP COLUMN IF EXISTS updated_by,
DROP COLUMN IF EXISTS created_by;

ALTER TABLE semesters
DROP COLUMN IF EXISTS updated_by,
DROP COLUMN IF EXISTS created_by;

-- +goose StatementEnd