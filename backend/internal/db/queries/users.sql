-- name: GetUser :one
SELECT id, email, password_hash, created_at, display_name, deletedAt
FROM users
WHERE id = $1
  AND deletedAt IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, display_name, deletedAt
FROM users
WHERE email = $1
  AND deletedAt IS NULL;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, display_name) VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, created_at, display_name, deletedAt;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
  AND deletedAt IS NULL
RETURNING id, email, password_hash, created_at, display_name, deletedAt;

-- name: DeleteUser :exec
UPDATE users SET deletedAt = NOW()
WHERE id = $1
  AND deletedAt IS NULL;

-- name: UpdateUserPasswordHash :exec
UPDATE users
SET password_hash = $2
WHERE id = $1
  AND deletedAt IS NULL;

-- name: ListForms :many
SELECT form_id, title FROM forms;

-- name: UpsertForm :exec
INSERT INTO forms (form_id, title, description, polling_sec)
VALUES ($1, $2, $3, $4)
ON CONFLICT (form_id) DO UPDATE
SET title = EXCLUDED.title, description = EXCLUDED.description, polling_sec = EXCLUDED.polling_sec;
