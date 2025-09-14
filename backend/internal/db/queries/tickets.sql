-- name: CreateTicket :one
INSERT INTO tickets (id, form_id, response_id, status, assignee_id)
VALUES ($1,$2,$3,'new',NULL)
RETURNING id, form_id, response_id, status, assignee_id, priority, created_at, updated_at;

-- name: ListTickets :many
SELECT id, form_id, response_id, status, assignee_id, priority, created_at, updated_at
FROM tickets
WHERE ($1::text IS NULL OR form_id = $1)
  AND ($2::ticket_status IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT 200;

-- name: GetTicket :one
SELECT id, form_id, response_id, status, assignee_id, priority, created_at, updated_at
FROM tickets WHERE id = $1;

-- name: UpdateTicket :one
UPDATE tickets
SET status = COALESCE($2,status),
    assignee_id = COALESCE($3,assignee_id),
    updated_at = NOW()
WHERE id = $1
RETURNING id, form_id, response_id, status, assignee_id, priority, created_at, updated_at;
