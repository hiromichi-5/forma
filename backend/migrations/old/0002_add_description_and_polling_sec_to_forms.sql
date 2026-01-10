-- +goose Up
ALTER TABLE forms ADD COLUMN description TEXT;
ALTER TABLE forms ADD COLUMN polling_sec INTEGER;

-- +goose Down
ALTER TABLE forms DROP COLUMN polling_sec;
ALTER TABLE forms DROP COLUMN description;
