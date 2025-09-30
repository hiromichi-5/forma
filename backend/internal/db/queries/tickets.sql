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
       r.payload
FROM tickets t
JOIN forms f ON f.form_id = t.form_id
JOIN responses r ON r.response_id = t.response_id
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
       r.payload
FROM tickets t
JOIN forms f ON f.form_id = t.form_id
JOIN responses r ON r.response_id = t.response_id
WHERE t.id = $1;

-- name: UpdateTicket :one
UPDATE tickets
SET status = COALESCE($2,status),
    assignee_id = COALESCE($3,assignee_id),
    priority = COALESCE($4,priority),
    updated_at = NOW()
WHERE id = $1
RETURNING id, form_id, response_id, status, assignee_id, priority, created_at, updated_at;
