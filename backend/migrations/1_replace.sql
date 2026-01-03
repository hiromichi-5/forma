-- +goose Up
CREATE TYPE form_role AS ENUM ('admin', 'editor');
CREATE TYPE ticket_priority AS ENUM ('high', 'medium', 'low');

CREATE TABLE users (
  id UUID PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  password_hash TEXT NOT NULL,
  display_name TEXT NOT NULL,
  verified_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE email_verification_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  token TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT email_verification_tokens_user_fk
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE password_reset_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL,
  token TEXT UNIQUE NOT NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT password_reset_tokens_user_fk
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE forms (
  id UUID PRIMARY KEY,
  form_id TEXT UNIQUE NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  title_question_id TEXT,
  email_collection_type TEXT,
  synced_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE form_members (
  user_id UUID NOT NULL,
  form_id UUID NOT NULL,
  role form_role NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (user_id, form_id),

  CONSTRAINT form_members_user_fk
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT form_members_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
);

CREATE TABLE form_invites (
  id UUID PRIMARY KEY,
  form_id UUID NOT NULL,
  email TEXT NOT NULL,
  role form_role NOT NULL,
  invited_by UUID NOT NULL,
  accepted_at TIMESTAMPTZ,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT form_invites_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE,
  CONSTRAINT form_invites_inviter_fk
    FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX form_invites_email_form_active
  ON form_invites(email, form_id)
  WHERE accepted_at IS NULL;

CREATE TABLE form_questions (
  form_id UUID NOT NULL,
  question_id TEXT NOT NULL,
  title TEXT NOT NULL,
  question_type TEXT NOT NULL,
  options JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  PRIMARY KEY (form_id, question_id),

  CONSTRAINT form_questions_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
);

ALTER TABLE forms
  ADD CONSTRAINT forms_title_question_fk
  FOREIGN KEY (id, title_question_id)
  REFERENCES form_questions(form_id, question_id);

CREATE TABLE form_statuses (
  id UUID PRIMARY KEY,
  form_id UUID NOT NULL,
  name TEXT NOT NULL,
  color TEXT,
  display_order INT NOT NULL,
  is_default BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT form_statuses_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX form_statuses_form_name
  ON form_statuses(form_id, name);

CREATE UNIQUE INDEX form_statuses_form_order
  ON form_statuses(form_id, display_order);

CREATE UNIQUE INDEX form_statuses_default
  ON form_statuses(form_id)
  WHERE is_default = TRUE;

CREATE TABLE tickets (
  id UUID PRIMARY KEY,
  form_id UUID NOT NULL,
  response_id TEXT NOT NULL,
  respondent_email TEXT,
  answers JSONB NOT NULL,
  status_id UUID NOT NULL,
  assignee_id UUID,
  priority ticket_priority NOT NULL DEFAULT 'medium',
  submitted_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT tickets_form_response_unique
    UNIQUE (form_id, response_id),
  CONSTRAINT tickets_form_fk
    FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE,
  CONSTRAINT tickets_status_fk
    FOREIGN KEY (status_id) REFERENCES form_statuses(id) ON DELETE RESTRICT,
  CONSTRAINT tickets_assignee_fk
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE ticket_histories (
  id UUID PRIMARY KEY,
  ticket_id UUID NOT NULL,
  changed_by UUID,
  changed_by_name TEXT NOT NULL,
  field_name TEXT NOT NULL,
  old_value TEXT,
  new_value TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ticket_histories_ticket_fk
    FOREIGN KEY (ticket_id) REFERENCES tickets(id) ON DELETE CASCADE,
  CONSTRAINT ticket_histories_user_fk
    FOREIGN KEY (changed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- +goose Down
DROP TABLE ticket_histories;
DROP TABLE tickets;
DROP TABLE form_statuses;
DROP TABLE form_questions;
DROP TABLE form_invites;
DROP TABLE form_members;
DROP TABLE forms;
DROP TABLE password_reset_tokens;
DROP TABLE email_verification_tokens;
DROP TABLE users;

DROP TYPE ticket_priority;
DROP TYPE form_role;
