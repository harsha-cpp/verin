-- name: InsertAuditEvent :one
INSERT INTO audit_events (
  org_id,
  actor_user_id,
  actor_role,
  action,
  resource_type,
  resource_id,
  ip_address,
  user_agent,
  request_id,
  payload_json
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListAuditEvents :many
SELECT * FROM audit_events
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAuditEventsFiltered :many
SELECT * FROM audit_events
WHERE org_id = $1
  AND ($4::text = '' OR action = $4)
  AND ($5::timestamptz IS NULL OR created_at >= $5)
  AND ($6::timestamptz IS NULL OR created_at <= $6)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
