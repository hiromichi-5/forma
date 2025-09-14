-- +goose Up
CREATE TABLE responses (
  response_id TEXT PRIMARY KEY,
  form_id TEXT NOT NULL REFERENCES forms(form_id) ON DELETE CASCADE,
  submitted_at TIMESTAMPTZ NOT NULL,
  payload JSONB NOT NULL,
  schema_version INT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_responses_form_time ON responses(form_id, submitted_at);

-- +goose Down
DROP INDEX IF EXISTS ix_responses_form_time;
DROP TABLE responses;
