-- name: GetUserFormRole :one
SELECT role FROM user_form_roles WHERE user_id = $1 AND form_id = $2;

-- name: UpsertUserFormRole :exec
INSERT INTO user_form_roles(user_id, form_id, role)
VALUES ($1,$2,$3)
ON CONFLICT (user_id, form_id) DO UPDATE SET role = EXCLUDED.role;

-- name: DeleteUserFormRole :exec
DELETE FROM user_form_roles WHERE user_id=$1 AND form_id=$2;

-- name: ListFormMembers :many
SELECT u.id, u.email, u.created_at, r.role
FROM user_form_roles r JOIN users u ON u.id = r.user_id
WHERE r.form_id = $1
ORDER BY u.email;
