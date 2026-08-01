-- name: CreateForm :one
INSERT INTO forms (id, form_id, title, description, title_question_id, email_collection_type, synced_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, form_id, title, description, title_question_id, email_collection_type, synced_at, created_at;

-- name: GetFormByID :one
SELECT id, form_id, title, description, title_question_id, email_collection_type, synced_at, created_at
FROM forms
WHERE id = $1;

-- name: UpdateFormTitleQuestion :execrows
UPDATE forms
SET title_question_id = $2
WHERE id = $1;

-- name: UpdateFormSyncedAt :execrows
UPDATE forms
SET synced_at = $2
WHERE id = $1;

-- name: DeleteForm :execrows
DELETE FROM forms
WHERE id = $1;

-- name: ListForms :many
SELECT id, form_id, title, description, title_question_id, email_collection_type, synced_at, created_at
FROM forms
ORDER BY created_at DESC;
