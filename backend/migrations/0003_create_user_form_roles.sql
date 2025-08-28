-- +goose Up
CREATE TABLE user_form_roles (
  user_id UUID NOT NULL,
  form_id TEXT NOT NULL,
  role TEXT NOT NULL,
  PRIMARY KEY (user_id, form_id)
);

-- +goose Down
DROP TABLE user_form_roles;
