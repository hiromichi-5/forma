-- name: GetUserByID :one
SELECT id, email, password_hash, created_at, display_name, verified_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, display_name, verified_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (id, email, password_hash, display_name, verified_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, email, password_hash, created_at, display_name, verified_at;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = $2
WHERE id = $1
RETURNING id, email, password_hash, created_at, display_name, verified_at;

-- name: UpdateUserPasswordHash :execrows
UPDATE users
SET password_hash = $2
WHERE id = $1;

-- name: SetUserVerifiedAt :execrows
UPDATE users
SET verified_at = $2
WHERE id = $1;

-- name: DeleteUser :execrows
DELETE FROM users
WHERE id = $1;
