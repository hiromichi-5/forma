-- +goose Up
ALTER TABLE forms
  ADD COLUMN title_question_id TEXT;

CREATE TABLE form_questions (
  form_id TEXT NOT NULL REFERENCES forms(form_id) ON DELETE CASCADE,
  question_id TEXT NOT NULL,
  title TEXT NOT NULL,
  question_type TEXT NOT NULL,
  options JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (form_id, question_id)
);

CREATE INDEX ix_form_questions_form ON form_questions(form_id);

-- +goose Down
DROP INDEX IF EXISTS ix_form_questions_form;
DROP TABLE IF EXISTS form_questions;
ALTER TABLE forms
  DROP COLUMN IF EXISTS title_question_id;
