-- name: ListQuotas :many
SELECT * FROM quotas
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpsertQuota :one
INSERT INTO quotas (
  org_id,
  target_type,
  target_id,
  max_storage_bytes,
  max_document_count
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (org_id, target_type, target_id)
DO UPDATE SET
  max_storage_bytes = EXCLUDED.max_storage_bytes,
  max_document_count = EXCLUDED.max_document_count,
  updated_at = now()
RETURNING *;

-- name: ListRetentionPolicies :many
SELECT * FROM retention_policies
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpsertRetentionPolicy :one
INSERT INTO retention_policies (
  id,
  org_id,
  name,
  applies_to_collection_id,
  retention_days,
  archive_after_days,
  delete_after_days,
  enabled
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (id)
DO UPDATE SET
  name = EXCLUDED.name,
  applies_to_collection_id = EXCLUDED.applies_to_collection_id,
  retention_days = EXCLUDED.retention_days,
  archive_after_days = EXCLUDED.archive_after_days,
  delete_after_days = EXCLUDED.delete_after_days,
  enabled = EXCLUDED.enabled,
  updated_at = now()
RETURNING *;

-- name: ListSystemSettings :many
SELECT * FROM system_settings
WHERE org_id = $1
ORDER BY setting_key;

-- name: UpsertSystemSetting :one
INSERT INTO system_settings (
  org_id,
  setting_key,
  setting_value
) VALUES ($1, $2, $3)
ON CONFLICT (org_id, setting_key)
DO UPDATE SET
  setting_value = EXCLUDED.setting_value,
  updated_at = now()
RETURNING *;

-- name: CreateJob :one
INSERT INTO ocr_jobs (
  document_version_id,
  job_type,
  task_id,
  status,
  payload_json
) VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateJobStatus :exec
UPDATE ocr_jobs
SET status = $2,
    attempt_count = $3,
    error_message = $4,
    updated_at = now(),
    finished_at = CASE WHEN $2 IN ('completed', 'failed') THEN now() ELSE finished_at END
WHERE task_id = $1;

-- name: ListJobs :many
SELECT * FROM ocr_jobs
ORDER BY created_at DESC
LIMIT $1;

-- name: GetJobByID :one
SELECT * FROM ocr_jobs
WHERE id = $1;

-- name: GetUsageSummary :one
SELECT
  COUNT(DISTINCT d.id)::bigint AS document_count,
  COALESCE(SUM(d.size_bytes), 0)::bigint AS storage_bytes,
  COUNT(DISTINCT u.id)::bigint AS user_count
FROM users u
LEFT JOIN documents d ON d.org_id = u.org_id AND d.deleted_at IS NULL
WHERE u.org_id = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations
ORDER BY name;
