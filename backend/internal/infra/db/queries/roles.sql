-- name: GetFormMemberRole :one
SELECT role
FROM form_members
WHERE user_id = $1
  AND form_id = $2;

-- name: UpsertFormMember :exec
INSERT INTO form_members (user_id, form_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, form_id) DO UPDATE
SET role = EXCLUDED.role;

-- name: DeleteFormMember :exec
DELETE FROM form_members
WHERE user_id = $1
  AND form_id = $2;

-- name: ListFormMembers :many
SELECT u.id, u.email, u.display_name, u.created_at, m.role
FROM form_members m
JOIN users u ON u.id = m.user_id
WHERE m.form_id = $1
ORDER BY u.email;

-- name: ListUserAccessibleForms :many
SELECT f.id, f.form_id, f.title, f.synced_at
FROM form_members m
JOIN forms f ON f.id = m.form_id
WHERE m.user_id = $1
ORDER BY f.title;

-- name: CountFormAdmins :one
SELECT COUNT(*)
FROM form_members
WHERE form_id = $1
  AND role = 'admin';
