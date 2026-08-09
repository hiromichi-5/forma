-- name: ListFormNotificationSettings :many
SELECT form_id, notification_type, mode, include_detail, updated_at
FROM form_notification_settings
WHERE form_id = $1;

-- name: GetFormNotificationSetting :one
SELECT form_id, notification_type, mode, include_detail, updated_at
FROM form_notification_settings
WHERE form_id = $1
  AND notification_type = $2;

-- name: UpsertFormNotificationSetting :one
INSERT INTO form_notification_settings (form_id, notification_type, mode, include_detail)
VALUES ($1, $2, $3, $4)
ON CONFLICT (form_id, notification_type)
DO UPDATE SET mode = EXCLUDED.mode,
              include_detail = EXCLUDED.include_detail,
              updated_at = NOW()
RETURNING form_id, notification_type, mode, include_detail, updated_at;

-- name: CreateTicketNotification :one
INSERT INTO ticket_notifications (id, ticket_id, notification_type, sent_by, sent_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, ticket_id, notification_type, sent_by, sent_at;

-- name: GetLatestTicketNotification :one
SELECT id, ticket_id, notification_type, sent_by, sent_at
FROM ticket_notifications
WHERE ticket_id = $1
  AND notification_type = $2
ORDER BY sent_at DESC
LIMIT 1;

-- name: ListLatestTicketNotifications :many
SELECT DISTINCT ON (notification_type)
       id, ticket_id, notification_type, sent_by, sent_at
FROM ticket_notifications
WHERE ticket_id = $1
ORDER BY notification_type, sent_at DESC;
