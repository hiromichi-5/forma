-- name: CreateSession :one
INSERT INTO sessions (id, user_id)
VALUES ($1, $2)
RETURNING id, user_id, created_at;

-- name: DeleteSession :execrows
DELETE FROM sessions
WHERE id = $1;

-- name: GetSessionByID :one
SELECT id, user_id, created_at
FROM sessions
WHERE id = $1;
