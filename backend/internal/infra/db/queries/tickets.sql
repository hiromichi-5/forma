-- name: CreateTicket :execrows
INSERT INTO tickets (id, form_id, response_id, respondent_email, answers, status_id, assignee_id, priority, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (form_id, response_id) DO NOTHING;

-- name: ListTickets :many
SELECT id, form_id, response_id, respondent_email, answers,
       status_id, assignee_id, priority, submitted_at, created_at
FROM tickets
WHERE form_id = $1
  AND ($2::uuid IS NULL OR status_id = $2)
ORDER BY created_at DESC
LIMIT 200;

-- name: GetTicket :one
SELECT id, form_id, response_id, respondent_email, answers,
       status_id, assignee_id, priority, submitted_at, created_at
FROM tickets
WHERE id = $1;

-- name: UpdateTicketStatus :exec
UPDATE tickets SET status_id = $2 WHERE id = $1;

-- name: UpdateTicketAssignee :exec
UPDATE tickets SET assignee_id = $2 WHERE id = $1;

-- name: UpdateTicketPriority :exec
UPDATE tickets SET priority = $2 WHERE id = $1;

-- name: CountTicketsByStatus :one
SELECT COUNT(1)
FROM tickets
WHERE status_id = $1;
