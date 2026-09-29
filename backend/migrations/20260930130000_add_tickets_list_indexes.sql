-- +goose Up
CREATE INDEX tickets_form_submitted_at_idx
  ON tickets(form_id, submitted_at DESC, id DESC);

CREATE INDEX tickets_form_status_submitted_at_idx
  ON tickets(form_id, status_id, submitted_at DESC, id DESC);

-- +goose Down
DROP INDEX tickets_form_status_submitted_at_idx;

DROP INDEX tickets_form_submitted_at_idx;
