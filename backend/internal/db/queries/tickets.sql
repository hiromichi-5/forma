-- name: CreateTicket :one
INSERT INTO tickets (id, form_id, response_id, status, assignee_id)
VALUES ($1,$2,$3,'new',NULL)
RETURNING id, form_id, response_id, status, assignee_id, priority, created_at, updated_at;

-- name: ListTickets :many
SELECT t.id,
       t.form_id,
       t.response_id,
       t.status,
       t.assignee_id,
       t.priority,
       t.created_at,
       t.updated_at,
       f.title AS form_title,
       f.title_question_id,
       r.submitted_at,
       r.payload,
       a.display_name AS assignee_display_name,
       a.email AS assignee_email
FROM tickets t
JOIN forms f ON f.form_id = t.form_id
JOIN responses r ON r.response_id = t.response_id
LEFT JOIN users a ON a.id = t.assignee_id
WHERE ($1::text IS NULL OR t.form_id = $1)
  AND ($2::text = '' OR $2::text IS NULL OR t.status = $2::ticket_status)
ORDER BY t.created_at DESC
LIMIT 200;

-- name: GetTicket :one
SELECT t.id,
       t.form_id,
       t.response_id,
       t.status,
       t.assignee_id,
       t.priority,
       t.created_at,
       t.updated_at,
       f.title AS form_title,
       f.title_question_id,
       r.submitted_at,
       r.payload,
       a.display_name AS assignee_display_name,
       a.email AS assignee_email
FROM tickets t
JOIN forms f ON f.form_id = t.form_id
JOIN responses r ON r.response_id = t.response_id
LEFT JOIN users a ON a.id = t.assignee_id
WHERE t.id = $1;

-- name: UpdateTicket :one
UPDATE tickets
SET status = COALESCE(sqlc.narg(status)::ticket_status, status),
    assignee_id = CASE
        WHEN sqlc.arg(clear_assignee)::bool THEN NULL
        WHEN sqlc.narg(assignee_id)::uuid IS NOT NULL THEN sqlc.narg(assignee_id)::uuid
        ELSE assignee_id
    END,
    priority = COALESCE(sqlc.narg(priority)::ticket_priority, priority),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING id, form_id, response_id, status, assignee_id, priority, created_at, updated_at;
