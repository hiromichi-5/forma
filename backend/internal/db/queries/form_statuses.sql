-- name: CreateFormStatus :one
INSERT INTO form_statuses (id, form_id, name, color, display_order, is_default)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, form_id, name, color, display_order, is_default, created_at;

-- name: ListFormStatuses :many
SELECT id, form_id, name, color, display_order, is_default, created_at
FROM form_statuses
WHERE form_id = $1
ORDER BY display_order;

-- name: GetDefaultFormStatus :one
SELECT id, form_id, name, color, display_order, is_default, created_at
FROM form_statuses
WHERE form_id = $1
  AND is_default = TRUE;

-- name: GetFormStatusByID :one
SELECT id, form_id, name, color, display_order, is_default, created_at
FROM form_statuses
WHERE id = $1;

-- name: ClearDefaultFormStatus :exec
UPDATE form_statuses
SET is_default = FALSE
WHERE form_id = $1
  AND is_default = TRUE;

-- name: SetDefaultFormStatus :one
UPDATE form_statuses AS fs
SET is_default = TRUE
WHERE fs.id = $2
  AND fs.form_id = $1
RETURNING fs.id, fs.form_id, fs.name, fs.color, fs.display_order, fs.is_default, fs.created_at;

-- name: UpdateFormStatus :one
UPDATE form_statuses
SET name = $2,
    color = $3,
    display_order = $4
WHERE id = $1
RETURNING id, form_id, name, color, display_order, is_default, created_at;

-- name: DeleteFormStatus :exec
DELETE FROM form_statuses
WHERE id = $1;
