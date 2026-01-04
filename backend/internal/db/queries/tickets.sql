-- name: CreateTicket :execrows
INSERT INTO tickets (id, form_id, response_id, respondent_email, answers, status_id, assignee_id, priority, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (form_id, response_id) DO NOTHING;

-- name: ListTickets :many
SELECT t.id,
       t.form_id,
       t.response_id,
       t.respondent_email,
       t.answers,
       t.status_id,
       t.assignee_id,
       t.priority,
       t.submitted_at,
       t.created_at,
       f.title AS form_title,
       f.title_question_id,
       s.name AS status_name,
       s.color AS status_color,
       a.display_name AS assignee_display_name,
       a.email AS assignee_email
FROM tickets t
JOIN forms f ON f.id = t.form_id
JOIN form_statuses s ON s.id = t.status_id
LEFT JOIN users a ON a.id = t.assignee_id
WHERE ($1::uuid IS NULL OR t.form_id = $1)
  AND ($2::uuid IS NULL OR t.status_id = $2)
ORDER BY t.created_at DESC
LIMIT 200;

-- name: GetTicket :one
SELECT t.id,
       t.form_id,
       t.response_id,
       t.respondent_email,
       t.answers,
       t.status_id,
       t.assignee_id,
       t.priority,
       t.submitted_at,
       t.created_at,
       f.title AS form_title,
       f.title_question_id,
       s.name AS status_name,
       s.color AS status_color,
       a.display_name AS assignee_display_name,
       a.email AS assignee_email
FROM tickets t
JOIN forms f ON f.id = t.form_id
JOIN form_statuses s ON s.id = t.status_id
LEFT JOIN users a ON a.id = t.assignee_id
WHERE t.id = $1;

-- name: UpdateTicketStatus :one
UPDATE tickets
SET status_id = $2
WHERE id = $1
RETURNING id, form_id, response_id, respondent_email, answers, status_id, assignee_id, priority, submitted_at, created_at;

-- name: UpdateTicketAssignee :one
UPDATE tickets
SET assignee_id = $2
WHERE id = $1
RETURNING id, form_id, response_id, respondent_email, answers, status_id, assignee_id, priority, submitted_at, created_at;

-- name: UpdateTicketPriority :one
UPDATE tickets
SET priority = $2
WHERE id = $1
RETURNING id, form_id, response_id, respondent_email, answers, status_id, assignee_id, priority, submitted_at, created_at;
