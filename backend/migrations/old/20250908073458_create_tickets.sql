-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE ticket_status AS ENUM ('new','in_progress','done');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

CREATE TABLE tickets (
  id UUID PRIMARY KEY,
  form_id TEXT NOT NULL REFERENCES forms(form_id) ON DELETE CASCADE,
  response_id TEXT NOT NULL REFERENCES responses(response_id) ON DELETE CASCADE,
  status ticket_status NOT NULL DEFAULT 'new',
  assignee_id UUID,
  priority INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_tickets_form_status ON tickets(form_id, status);

-- +goose Down
DROP INDEX IF EXISTS ix_tickets_form_status;
DROP TABLE tickets;
DROP TYPE IF EXISTS ticket_status;
