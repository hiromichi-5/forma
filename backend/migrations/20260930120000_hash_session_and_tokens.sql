-- +goose Up
-- 既存の Cookie（セッション ID の UUID 文字列）がそのまま使えるよう、同じ値のハッシュを移す
ALTER TABLE sessions ADD COLUMN token_hash BYTEA;

UPDATE sessions SET token_hash = sha256(convert_to(id::text, 'UTF8'));

ALTER TABLE sessions ALTER COLUMN token_hash SET NOT NULL;
ALTER TABLE sessions ADD CONSTRAINT sessions_token_hash_key UNIQUE (token_hash);

ALTER TABLE email_verification_tokens ADD COLUMN token_hash BYTEA;

UPDATE email_verification_tokens SET token_hash = sha256(convert_to(token, 'UTF8'));

ALTER TABLE email_verification_tokens ALTER COLUMN token_hash SET NOT NULL;
ALTER TABLE email_verification_tokens
  ADD CONSTRAINT email_verification_tokens_token_hash_key UNIQUE (token_hash);
ALTER TABLE email_verification_tokens DROP COLUMN token;

ALTER TABLE password_reset_tokens ADD COLUMN token_hash BYTEA;

UPDATE password_reset_tokens SET token_hash = sha256(convert_to(token, 'UTF8'));

ALTER TABLE password_reset_tokens ALTER COLUMN token_hash SET NOT NULL;
ALTER TABLE password_reset_tokens
  ADD CONSTRAINT password_reset_tokens_token_hash_key UNIQUE (token_hash);
ALTER TABLE password_reset_tokens DROP COLUMN token;

-- +goose Down
-- ハッシュから平文は復元できないため、未使用のトークンは破棄する
DELETE FROM password_reset_tokens;
ALTER TABLE password_reset_tokens DROP COLUMN token_hash;
ALTER TABLE password_reset_tokens ADD COLUMN token TEXT UNIQUE NOT NULL;

DELETE FROM email_verification_tokens;
ALTER TABLE email_verification_tokens DROP COLUMN token_hash;
ALTER TABLE email_verification_tokens ADD COLUMN token TEXT UNIQUE NOT NULL;

ALTER TABLE sessions DROP COLUMN token_hash;
