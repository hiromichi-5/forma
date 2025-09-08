-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at FROM users WHERE email=$1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)
RETURNING id, email, password_hash, created_at;

-- name: ListForms :many
SELECT form_id, title FROM forms;

-- name: UpsertForm :exec
INSERT INTO forms (form_id, title, description, polling_sec)
VALUES ($1, $2, $3, $4)
ON CONFLICT (form_id) DO UPDATE
SET title = EXCLUDED.title, description = EXCLUDED.description, polling_sec = EXCLUDED.polling_sec;
