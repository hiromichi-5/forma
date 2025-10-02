-- name: UpdateFormTitleQuestion :exec
UPDATE forms
SET title_question_id = $2
WHERE form_id = $1;

-- name: GetFormTitleQuestion :one
SELECT title_question_id FROM forms WHERE form_id = $1;
