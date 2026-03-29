-- name: CreateTicketHistory :one
INSERT INTO ticket_histories (id, ticket_id, changed_by, changed_by_name, field_name, old_value, new_value)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, ticket_id, changed_by, changed_by_name, field_name, old_value, new_value, created_at;

-- name: ListTicketHistoriesByTicket :many
SELECT id, ticket_id, changed_by, changed_by_name, field_name, old_value, new_value, created_at
FROM ticket_histories
WHERE ticket_id = $1
ORDER BY created_at DESC;
