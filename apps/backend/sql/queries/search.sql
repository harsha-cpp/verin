-- name: SearchDocuments :many
SELECT DISTINCT
  d.id,
  d.title,
  d.original_filename,
  d.status,
  d.updated_at,
  COALESCE(ts_rank(d.search_vector, plainto_tsquery('simple', $5)), 0)
    + COALESCE(ts_rank(ot.search_vector, plainto_tsquery('simple', $5)), 0) AS rank,
  COALESCE(LEFT(ot.content, 200), '') AS snippet
FROM documents d
LEFT JOIN document_versions dv ON dv.id = d.current_version_id
LEFT JOIN ocr_text ot ON ot.document_version_id = dv.id
WHERE d.org_id = $1
  AND d.deleted_at IS NULL
  AND (
    $4::boolean
    OR d.owner_user_id = $2
    OR EXISTS (
      SELECT 1
      FROM collection_members cm
      WHERE cm.collection_id = d.collection_id
        AND cm.user_id = $2
    )
    OR EXISTS (
      SELECT 1
      FROM document_permissions dp
      WHERE dp.document_id = d.id
        AND dp.allow = true
        AND (
          (dp.subject_type = 'user' AND dp.subject_id = $2)
          OR (dp.subject_type = 'role' AND dp.subject_id = ANY($3::uuid[]))
        )
    )
  )
  AND (
    d.search_vector @@ plainto_tsquery('simple', $5)
    OR ot.search_vector @@ plainto_tsquery('simple', $5)
    OR EXISTS (
      SELECT 1
      FROM document_metadata dm
      WHERE dm.document_id = d.id
        AND coalesce(dm.value_text, '') ILIKE '%' || $5 || '%'
    )
  )
ORDER BY rank DESC, d.updated_at DESC
LIMIT $6 OFFSET $7;

-- name: SearchDocumentsWithFilters :many
SELECT DISTINCT
  d.id,
  d.title,
  d.original_filename,
  d.status,
  d.mime_type,
  d.updated_at,
  COALESCE(ts_rank(d.search_vector, plainto_tsquery('simple', $5)), 0)
    + COALESCE(ts_rank(ot.search_vector, plainto_tsquery('simple', $5)), 0) AS rank,
  COALESCE(LEFT(ot.content, 200), '') AS snippet
FROM documents d
LEFT JOIN document_versions dv ON dv.id = d.current_version_id
LEFT JOIN ocr_text ot ON ot.document_version_id = dv.id
WHERE d.org_id = $1
  AND d.deleted_at IS NULL
  AND (
    $4::boolean
    OR d.owner_user_id = $2
    OR EXISTS (
      SELECT 1 FROM collection_members cm
      WHERE cm.collection_id = d.collection_id AND cm.user_id = $2
    )
    OR EXISTS (
      SELECT 1 FROM document_permissions dp
      WHERE dp.document_id = d.id AND dp.allow = true
        AND ((dp.subject_type = 'user' AND dp.subject_id = $2) OR (dp.subject_type = 'role' AND dp.subject_id = ANY($3::uuid[])))
    )
  )
  AND (
    $5::text = ''
    OR d.search_vector @@ plainto_tsquery('simple', $5)
    OR ot.search_vector @@ plainto_tsquery('simple', $5)
  )
  AND ($8::text = '' OR d.status = $8)
  AND ($9::text = '' OR d.mime_type = $9)
  AND ($10::timestamptz IS NULL OR d.updated_at >= $10)
  AND ($11::timestamptz IS NULL OR d.updated_at <= $11)
ORDER BY rank DESC, d.updated_at DESC
LIMIT $6 OFFSET $7;

-- name: ListSavedSearches :many
SELECT * FROM saved_searches
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: CreateSavedSearch :one
INSERT INTO saved_searches (
  org_id,
  user_id,
  name,
  query_text,
  filters_json
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteSavedSearch :exec
DELETE FROM saved_searches
WHERE id = $1
  AND user_id = $2;
