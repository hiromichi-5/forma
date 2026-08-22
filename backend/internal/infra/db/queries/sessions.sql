-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, created_at, expires_at;

-- name: DeleteSession :execrows
DELETE FROM sessions
WHERE id = $1;

-- name: GetSessionByID :one
SELECT id, user_id, created_at, expires_at
FROM sessions
WHERE id = $1
  AND expires_at > NOW();
