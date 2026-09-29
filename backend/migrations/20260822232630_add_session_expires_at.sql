-- +goose Up
ALTER TABLE sessions ADD COLUMN expires_at TIMESTAMPTZ;

UPDATE sessions SET expires_at = created_at + INTERVAL '14 days';

ALTER TABLE sessions ALTER COLUMN expires_at SET NOT NULL;

CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);

-- +goose Down
DROP INDEX sessions_expires_at_idx;

ALTER TABLE sessions DROP COLUMN expires_at;
