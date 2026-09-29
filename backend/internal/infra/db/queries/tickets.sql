-- name: CreateTicket :execrows
INSERT INTO tickets (id, form_id, response_id, respondent_email, answers, status_id, assignee_id, priority, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (form_id, response_id) DO NOTHING;

-- name: ListTickets :many
SELECT id, form_id, response_id, respondent_email, answers,
       status_id, assignee_id, priority, submitted_at, created_at
FROM tickets
WHERE form_id = sqlc.arg(form_id)
  AND (COALESCE(cardinality(sqlc.arg(status_ids)::uuid[]), 0) = 0 OR status_id = ANY(sqlc.arg(status_ids)::uuid[]))
  AND (sqlc.narg(email_pattern)::text IS NULL OR respondent_email ILIKE sqlc.narg(email_pattern))
  AND (sqlc.narg(cursor_submitted_at)::timestamptz IS NULL
       OR (submitted_at, id) < (sqlc.narg(cursor_submitted_at)::timestamptz, sqlc.narg(cursor_id)::uuid))
ORDER BY submitted_at DESC, id DESC
LIMIT sqlc.arg(row_limit);

-- name: CountTicketsGroupByStatus :many
SELECT status_id, COUNT(1) AS count
FROM tickets
WHERE form_id = sqlc.arg(form_id)
  AND (sqlc.narg(email_pattern)::text IS NULL OR respondent_email ILIKE sqlc.narg(email_pattern))
GROUP BY status_id;

-- name: SummarizeTicketsByForms :many
SELECT form_id, COUNT(1) AS ticket_count, MAX(submitted_at)::timestamptz AS latest_submitted_at
FROM tickets
WHERE form_id = ANY(sqlc.arg(form_ids)::uuid[])
GROUP BY form_id;

-- name: GetTicket :one
SELECT id, form_id, response_id, respondent_email, answers,
       status_id, assignee_id, priority, submitted_at, created_at
FROM tickets
WHERE id = $1;

-- name: UpdateTicket :execrows
UPDATE tickets
SET status_id = $2,
    assignee_id = $3,
    priority = $4
WHERE id = $1;

-- name: CountTicketsByStatus :one
SELECT COUNT(1)
FROM tickets
WHERE status_id = $1;
