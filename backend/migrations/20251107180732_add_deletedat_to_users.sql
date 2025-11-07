-- +goose Up
-- +goose StatementBegin
-- Add deletedAt column to users table
ALTER TABLE users
ADD COLUMN deletedAt TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
DROP COLUMN deletedAt;
-- +goose StatementEnd
