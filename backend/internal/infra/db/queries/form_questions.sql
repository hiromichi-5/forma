-- name: UpsertFormQuestion :exec
INSERT INTO form_questions (form_id, question_id, title, question_type, options)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (form_id, question_id) DO UPDATE
SET title = EXCLUDED.title,
    question_type = EXCLUDED.question_type,
    options = EXCLUDED.options;

-- name: ListFormQuestions :many
SELECT form_id, question_id, title, question_type, options, created_at
FROM form_questions
WHERE form_id = $1
ORDER BY title;

-- name: DeleteFormQuestion :exec
DELETE FROM form_questions
WHERE form_id = $1
  AND question_id = $2;
