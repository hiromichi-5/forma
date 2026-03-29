-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (id, user_id, token, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token, expires_at, used_at, created_at;

-- name: GetPasswordResetTokenByToken :one
SELECT id, user_id, token, expires_at, used_at, created_at
FROM password_reset_tokens
WHERE token = $1
  AND used_at IS NULL
  AND expires_at > NOW();

-- name: UsePasswordResetToken :execrows
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE id = $1
  AND used_at IS NULL;

-- name: DeletePasswordResetTokensByUser :exec
DELETE FROM password_reset_tokens
WHERE user_id = $1;
