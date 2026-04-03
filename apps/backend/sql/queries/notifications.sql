-- name: ListNotificationsByUserID :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: CreateNotification :one
INSERT INTO notifications (
  org_id,
  user_id,
  kind,
  title,
  body,
  payload_json
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET read_at = now()
WHERE id = $1
  AND user_id = $2;
