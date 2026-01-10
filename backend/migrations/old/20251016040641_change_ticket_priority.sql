-- +goose Up
-- +goose StatementBegin
DO $$ BEGIN
  CREATE TYPE ticket_priority AS ENUM ('High', 'Medium', 'Low');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;
-- +goose StatementEnd

ALTER TABLE tickets
  ADD COLUMN priority_tmp ticket_priority;

UPDATE tickets
SET priority_tmp = CASE
  WHEN priority >= 3 THEN 'High'::ticket_priority
  WHEN priority >= 2 THEN 'Medium'::ticket_priority
  ELSE 'Low'::ticket_priority
END;

ALTER TABLE tickets DROP COLUMN priority;

ALTER TABLE tickets
  RENAME COLUMN priority_tmp TO priority;

ALTER TABLE tickets
  ALTER COLUMN priority SET DEFAULT 'Medium'::ticket_priority,
  ALTER COLUMN priority SET NOT NULL;

-- +goose Down
ALTER TABLE tickets
  ADD COLUMN priority_tmp INT;

UPDATE tickets
SET priority_tmp = CASE
  WHEN priority = 'High'::ticket_priority THEN 3
  WHEN priority = 'Medium'::ticket_priority THEN 2
  ELSE 0
END;

ALTER TABLE tickets DROP COLUMN priority;
ALTER TABLE tickets RENAME COLUMN priority_tmp TO priority;

ALTER TABLE tickets
  ALTER COLUMN priority SET DEFAULT 0,
  ALTER COLUMN priority SET NOT NULL;

-- +goose StatementBegin
DO $$ BEGIN
  DROP TYPE ticket_priority;
EXCEPTION WHEN undefined_object THEN NULL; END $$;
-- +goose StatementEnd
