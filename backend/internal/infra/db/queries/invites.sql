-- name: CreateFormInvite :one
INSERT INTO form_invites (id, form_id, email, role, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, form_id, email, role, invited_by, accepted_at, expires_at, created_at;

-- name: ListActiveFormInvites :many
SELECT id, form_id, email, role, invited_by, accepted_at, expires_at, created_at
FROM form_invites
WHERE form_id = $1
  AND accepted_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: GetFormInviteByEmail :one
SELECT id, form_id, email, role, invited_by, accepted_at, expires_at, created_at
FROM form_invites
WHERE form_id = $1
  AND email = $2
  AND accepted_at IS NULL;

-- name: GetFormInviteForUpdate :one
SELECT id, form_id, email, role, invited_by, accepted_at, expires_at, created_at
FROM form_invites
WHERE id = $1
FOR UPDATE;

-- name: AcceptFormInvite :one
UPDATE form_invites
SET accepted_at = NOW()
WHERE id = $1
  AND accepted_at IS NULL
RETURNING id, form_id, email, role, invited_by, accepted_at, expires_at, created_at;

-- name: DeleteFormInvite :execrows
DELETE FROM form_invites
WHERE id = $1;
