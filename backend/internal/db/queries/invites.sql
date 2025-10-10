-- name: CreateFormInvite :one
INSERT INTO form_invites (code, form_id, role, expires_at, created_by)
VALUES ($1, $2, $3, $4, $5)
RETURNING code, form_id, role, expires_at, created_by, created_at, revoked;

-- name: ListActiveFormInvites :many
SELECT code, form_id, role, expires_at, created_by, created_at, revoked
FROM form_invites
WHERE form_id = $1
  AND revoked = FALSE
  AND expires_at > $2
ORDER BY created_at DESC;

-- name: GetFormInviteForUpdate :one
SELECT code, form_id, role, expires_at, created_by, created_at, revoked
FROM form_invites
WHERE code = $1
FOR UPDATE;

-- name: RevokeFormInvite :one
UPDATE form_invites
SET revoked = TRUE
WHERE form_id = $1
  AND code = $2
RETURNING code, form_id, role, expires_at, created_by, created_at, revoked;
