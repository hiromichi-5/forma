-- name: UpdateSyncCursor :exec
UPDATE forms SET sync_cursor = $2 WHERE form_id = $1;

-- name: GetFormSyncCursor :one
SELECT sync_cursor FROM forms WHERE form_id = $1;
