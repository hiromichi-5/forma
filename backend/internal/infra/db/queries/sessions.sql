-- name: CreateSession :one
INSERT INTO sessions (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, created_at, expires_at, token_hash;

-- name: DeleteSessionByTokenHash :execrows
DELETE FROM sessions
WHERE token_hash = $1;

-- name: DeleteSessionsByUser :exec
DELETE FROM sessions
WHERE user_id = $1;

-- name: DeleteSessionsByUserExcept :exec
DELETE FROM sessions
WHERE user_id = $1
  AND id <> $2;

-- name: GetSessionByTokenHash :one
SELECT id, user_id, created_at, expires_at, token_hash
FROM sessions
WHERE token_hash = $1
  AND expires_at > NOW();
