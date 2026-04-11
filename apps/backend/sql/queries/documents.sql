-- name: ListCollections :many
SELECT * FROM collections
WHERE org_id = $1
ORDER BY name;

-- name: GetCollectionByID :one
SELECT * FROM collections
WHERE id = $1;

-- name: IsCollectionMember :one
SELECT EXISTS (
  SELECT 1
  FROM collection_members
  WHERE collection_id = $1
    AND user_id = $2
);

-- name: ListAccessibleDocuments :many
SELECT DISTINCT
  d.*,
  dv.version_number AS current_version_number,
  COALESCE(pa.storage_key, '') AS preview_storage_key
FROM documents d
LEFT JOIN document_versions dv ON dv.id = d.current_version_id
LEFT JOIN preview_assets pa
  ON pa.document_version_id = d.current_version_id
 AND pa.asset_type = 'preview'
WHERE d.org_id = $1
  AND d.deleted_at IS NULL
  AND (
    $4::boolean
    OR d.owner_user_id = $2
    OR EXISTS (
      SELECT 1 FROM users u WHERE u.id = $2 AND u.org_id = $1
    )
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
    $5::text = ''
    OR d.title ILIKE '%' || $5 || '%'
    OR d.original_filename ILIKE '%' || $5 || '%'
  )
ORDER BY d.updated_at DESC
LIMIT $6 OFFSET $7;

-- name: CountAccessibleDocuments :one
SELECT count(DISTINCT d.id)
FROM documents d
WHERE d.org_id = $1
  AND d.deleted_at IS NULL
  AND (
    $4::boolean
    OR d.owner_user_id = $2
    OR EXISTS (
      SELECT 1 FROM users u WHERE u.id = $2 AND u.org_id = $1
    )
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
    $5::text = ''
    OR d.title ILIKE '%' || $5 || '%'
    OR d.original_filename ILIKE '%' || $5 || '%'
  );

-- name: GetDocumentByID :one
SELECT * FROM documents
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetDocumentByVersionID :one
SELECT d.*
FROM documents d
JOIN document_versions dv ON dv.document_id = d.id
WHERE dv.id = $1
  AND d.deleted_at IS NULL;

-- name: CreateUpload :one
INSERT INTO uploads (
  org_id,
  user_id,
  object_key,
  original_filename,
  mime_type,
  size_bytes,
  checksum_sha256,
  status,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetUploadByID :one
SELECT * FROM uploads
WHERE id = $1;

-- name: CompleteUpload :exec
UPDATE uploads
SET status = $2,
    document_id = $3,
    completed_at = now()
WHERE id = $1;

-- name: CreateDocument :one
INSERT INTO documents (
  id,
  org_id,
  title,
  original_filename,
  mime_type,
  size_bytes,
  checksum_sha256,
  owner_user_id,
  collection_id,
  status
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateDocumentCore :one
UPDATE documents
SET title = $2,
    collection_id = $3,
    status = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetDocumentCurrentVersion :exec
UPDATE documents
SET current_version_id = $2,
    size_bytes = $3,
    mime_type = $4,
    checksum_sha256 = $5,
    updated_at = now()
WHERE id = $1;

-- name: ArchiveDocument :exec
UPDATE documents
SET status = 'archived',
    archived_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: RestoreDocument :exec
UPDATE documents
SET status = 'ready',
    archived_at = NULL,
    updated_at = now()
WHERE id = $1;

-- name: SetDocumentStatus :exec
UPDATE documents
SET status = $2,
    updated_at = now()
WHERE id = $1;

-- name: CreateDocumentVersion :one
INSERT INTO document_versions (
  id,
  document_id,
  version_number,
  storage_key,
  size_bytes,
  checksum_sha256,
  mime_type,
  uploaded_by,
  change_summary
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetLatestVersionNumber :one
SELECT COALESCE(MAX(version_number), 0)::int4
FROM document_versions
WHERE document_id = $1;

-- name: ListDocumentVersions :many
SELECT * FROM document_versions
WHERE document_id = $1
ORDER BY version_number DESC;

-- name: GetDocumentVersionByID :one
SELECT * FROM document_versions
WHERE id = $1;

-- name: DeleteMetadataForDocument :exec
DELETE FROM document_metadata
WHERE document_id = $1;

-- name: InsertMetadata :exec
INSERT INTO document_metadata (
  document_id,
  schema_key,
  value_text,
  value_number,
  value_boolean,
  value_date,
  value_jsonb
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
);

-- name: ListMetadataByDocumentID :many
SELECT * FROM document_metadata
WHERE document_id = $1
ORDER BY schema_key;

-- name: UpsertTag :one
INSERT INTO tags (org_id, name, color)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, name)
DO UPDATE SET color = EXCLUDED.color
RETURNING *;

-- name: DeleteDocumentTags :exec
DELETE FROM document_tags
WHERE document_id = $1;

-- name: AttachTag :exec
INSERT INTO document_tags (document_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListTagsByDocumentID :many
SELECT t.*
FROM tags t
JOIN document_tags dt ON dt.tag_id = t.id
WHERE dt.document_id = $1
ORDER BY t.name;

-- name: InsertDocumentPermission :exec
INSERT INTO document_permissions (
  document_id,
  subject_type,
  subject_id,
  access_level,
  allow
) VALUES ($1, $2, $3, $4, $5);

-- name: ListDocumentPermissions :many
SELECT * FROM document_permissions
WHERE document_id = $1
ORDER BY created_at DESC;

-- name: UpsertPreviewAsset :exec
INSERT INTO preview_assets (
  document_version_id,
  asset_type,
  storage_key,
  status
) VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: ListPreviewAssetsByVersionID :many
SELECT * FROM preview_assets
WHERE document_version_id = $1
ORDER BY created_at DESC;

-- name: GetOCRTextByVersionID :one
SELECT * FROM ocr_text
WHERE document_version_id = $1;

-- name: UpsertOCRText :one
INSERT INTO ocr_text (
  document_version_id,
  content,
  extraction_status,
  extracted_at,
  language,
  confidence_score
) VALUES (
  $1, $2, $3, $4, $5, $6
)
ON CONFLICT (document_version_id)
DO UPDATE SET
  content = EXCLUDED.content,
  extraction_status = EXCLUDED.extraction_status,
  extracted_at = EXCLUDED.extracted_at,
  language = EXCLUDED.language,
  confidence_score = EXCLUDED.confidence_score
RETURNING *;

-- name: SoftDeleteDocument :exec
UPDATE documents
SET status = 'deleted',
    deleted_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: CreateCollection :one
INSERT INTO collections (org_id, name, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateCollection :one
UPDATE collections
SET name = $2,
    description = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteCollection :exec
DELETE FROM collections
WHERE id = $1;

-- name: AddCollectionMember :exec
INSERT INTO collection_members (collection_id, user_id, access_level)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: RemoveCollectionMember :exec
DELETE FROM collection_members
WHERE collection_id = $1
  AND user_id = $2;

-- name: ListCollectionMembers :many
SELECT cm.*, u.full_name, u.email
FROM collection_members cm
JOIN users u ON u.id = cm.user_id
WHERE cm.collection_id = $1
ORDER BY u.full_name;

-- name: ListDocumentsByCollectionID :many
SELECT d.*, dv.version_number AS current_version_number
FROM documents d
LEFT JOIN document_versions dv ON dv.id = d.current_version_id
WHERE d.collection_id = $1
  AND d.deleted_at IS NULL
ORDER BY d.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: ShareDocument :exec
INSERT INTO document_permissions (document_id, subject_type, subject_id, access_level, allow)
VALUES ($1, $2, $3, $4, true)
ON CONFLICT DO NOTHING;

-- name: RevokeDocumentShare :exec
DELETE FROM document_permissions
WHERE document_id = $1
  AND subject_type = $2
  AND subject_id = $3;

-- name: ListSharedDocuments :many
SELECT DISTINCT d.*, dv.version_number AS current_version_number
FROM documents d
JOIN document_permissions dp ON dp.document_id = d.id
LEFT JOIN document_versions dv ON dv.id = d.current_version_id
WHERE dp.subject_type = 'user'
  AND dp.subject_id = $1
  AND dp.allow = true
  AND d.deleted_at IS NULL
ORDER BY d.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrphanUploads :many
SELECT * FROM uploads
WHERE status = 'pending'
  AND expires_at < now()
ORDER BY created_at ASC
LIMIT $1;

-- name: DeleteUpload :exec
DELETE FROM uploads
WHERE id = $1;

-- name: ArchiveExpiredDocuments :exec
UPDATE documents
SET status = 'archived',
    archived_at = now(),
    updated_at = now()
WHERE org_id = $1
  AND deleted_at IS NULL
  AND status = 'ready'
  AND updated_at < $2;
