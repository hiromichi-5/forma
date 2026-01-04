-- name: CreateEmailVerificationToken :one
INSERT INTO email_verification_tokens (id, user_id, token, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token, expires_at, used_at, created_at;

-- name: GetEmailVerificationTokenByToken :one
SELECT id, user_id, token, expires_at, used_at, created_at
FROM email_verification_tokens
WHERE token = $1
  AND used_at IS NULL
  AND expires_at > NOW();

-- name: UseEmailVerificationToken :execrows
UPDATE email_verification_tokens
SET used_at = NOW()
WHERE id = $1
  AND used_at IS NULL;

-- name: DeleteEmailVerificationTokensByUser :exec
DELETE FROM email_verification_tokens
WHERE user_id = $1;
