-- +goose Up
ALTER TABLE forms ADD COLUMN sync_cursor TIMESTAMPTZ;
ALTER TABLE forms ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE forms DROP COLUMN enabled;
ALTER TABLE forms DROP COLUMN sync_cursor;
