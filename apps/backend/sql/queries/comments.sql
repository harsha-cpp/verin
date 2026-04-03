-- name: ListCommentsByDocumentID :many
SELECT
  c.*,
  u.full_name AS author_name
FROM comments c
JOIN users u ON u.id = c.author_user_id
WHERE c.document_id = $1
ORDER BY c.created_at DESC;

-- name: CreateComment :one
INSERT INTO comments (
  document_id,
  author_user_id,
  body
) VALUES ($1, $2, $3)
RETURNING *;
