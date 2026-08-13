-- +goose Up
CREATE TYPE notification_type AS ENUM ('status_change', 'assignee_assigned');
CREATE TYPE notification_mode AS ENUM ('always', 'confirm', 'off');

CREATE TABLE form_notification_settings (
  form_id UUID NOT NULL,
  notification_type notification_type NOT NULL,
  mode notification_mode NOT NULL DEFAULT 'off',
  include_detail BOOLEAN NOT NULL DEFAULT FALSE,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (form_id, notification_type),

  CONSTRAINT form_notification_settings_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
);

CREATE TABLE ticket_notifications (
  id UUID PRIMARY KEY,
  ticket_id UUID NOT NULL,
  notification_type notification_type NOT NULL,
  sent_by UUID,
  sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ticket_notifications_ticket_fk
    FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
  CONSTRAINT ticket_notifications_user_fk
    FOREIGN KEY (sent_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX ticket_notifications_ticket_type_sent_at
  ON ticket_notifications(ticket_id, notification_type, sent_at DESC);

-- +goose Down
DROP TABLE ticket_notifications;
DROP TABLE form_notification_settings;

DROP TYPE notification_mode;
DROP TYPE notification_type;
