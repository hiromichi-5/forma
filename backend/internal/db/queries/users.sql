-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at FROM users WHERE email=$1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)
RETURNING id, email, password_hash, created_at;
