-- +goose Up
CREATE TABLE form_invites (
  code TEXT PRIMARY KEY,
  form_id TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'editor',
  expires_at TIMESTAMPTZ NOT NULL,
  created_by UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  revoked BOOLEAN NOT NULL DEFAULT FALSE,
  CONSTRAINT form_invites_role_check CHECK (role IN ('admin', 'editor')),
  -- 存在しないフォームの招待コードを作成しない、フォームが削除された時に招待コードを削除する
  CONSTRAINT form_invites_form_fk FOREIGN KEY (form_id) REFERENCES forms(form_id) ON DELETE CASCADE,
  -- 招待コードを発行したユーザーが削除されたときにそのコードを削除する
  CONSTRAINT form_invites_creator_fk FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX form_invites_form_id_idx ON form_invites (form_id);
CREATE INDEX form_invites_active_idx ON form_invites (form_id) WHERE revoked = FALSE;

-- +goose Down
DROP TABLE form_invites;
