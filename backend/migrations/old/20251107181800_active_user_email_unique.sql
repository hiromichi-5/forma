-- +goose Up
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX users_email_active_unique ON users (email) WHERE deletedAt IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_email_active_unique;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
-- +goose StatementEnd
