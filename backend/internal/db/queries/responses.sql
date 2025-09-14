-- name: InsertResponse :execrows
INSERT INTO responses (response_id, form_id, submitted_at, payload, schema_version)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (response_id) DO NOTHING;

-- name: ListResponses :many
SELECT response_id, form_id, submitted_at, payload, schema_version, created_at
FROM responses
WHERE ($1::text IS NULL OR form_id = $1)
  AND ($2::timestamptz IS NULL OR submitted_at >= $2)
ORDER BY submitted_at DESC
LIMIT 200;

-- name: GetResponseExists :one
SELECT EXISTS(SELECT 1 FROM responses WHERE response_id = $1) AS exists;
