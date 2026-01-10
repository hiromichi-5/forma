-- +goose Up
-- 既存のNULL値にデフォルト値を設定
UPDATE users SET display_name = email WHERE display_name IS NULL;

-- NOT NULL制約を追加
ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;

-- +goose Down
ALTER TABLE users ALTER COLUMN display_name DROP NOT NULL;