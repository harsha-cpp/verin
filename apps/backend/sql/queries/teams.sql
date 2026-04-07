-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1;

-- name: CreateGoogleUser :one
INSERT INTO users (email, full_name, google_id, avatar_url, password_hash, status)
VALUES ($1, $2, $3, $4, '', 'active')
RETURNING *;

-- name: SetUserOrg :exec
UPDATE users
SET org_id = $2, updated_at = now()
WHERE id = $1;

-- name: CreateOrganization :one
INSERT INTO organizations (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations
WHERE id = $1;

-- name: CreateRole :one
INSERT INTO roles (org_id, key, name, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPermissionByKey :one
SELECT * FROM permissions
WHERE key = $1;

-- name: ListAllPermissions :many
SELECT * FROM permissions
ORDER BY key;

-- name: AddRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: CreateTeamInvite :one
INSERT INTO team_invites (org_id, code, created_by, expires_at, max_uses)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ClaimTeamInvite :one
UPDATE team_invites
SET use_count = use_count + 1
WHERE code = $1
  AND (max_uses = 0 OR use_count < max_uses)
  AND expires_at > now()
RETURNING *;

-- name: GetTeamInviteByCode :one
SELECT * FROM team_invites
WHERE code = $1;

-- name: ListTeamMembers :many
SELECT
  u.id,
  u.email,
  u.full_name,
  u.avatar_url,
  u.status,
  u.created_at,
  COALESCE(array_remove(array_agg(r.key), NULL), ARRAY[]::text[])::text[] AS role_keys
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE u.org_id = $1
GROUP BY u.id
ORDER BY u.created_at;

-- name: CreateDefaultQuota :one
INSERT INTO quotas (org_id, target_type, max_storage_bytes, max_document_count)
VALUES ($1, 'org', 10737418240, 50000)
RETURNING *;

-- name: EnsureGlobalPermissions :exec
INSERT INTO permissions (id, key, description)
VALUES
  ('20000000-0000-0000-0000-000000000001', 'documents.read', 'Read documents'),
  ('20000000-0000-0000-0000-000000000002', 'documents.write', 'Create and update documents'),
  ('20000000-0000-0000-0000-000000000003', 'documents.admin', 'Archive and restore documents'),
  ('20000000-0000-0000-0000-000000000004', 'audit.read', 'Read audit logs'),
  ('20000000-0000-0000-0000-000000000005', 'admin.manage', 'Manage users and settings')
ON CONFLICT (id) DO NOTHING;

-- name: GetRoleByOrgAndKey :one
SELECT * FROM roles
WHERE org_id = $1 AND key = $2;

-- name: UpdateMemberRole :exec
UPDATE user_roles
SET role_id = $2
WHERE user_id = $1;

-- name: RemoveTeamMember :exec
UPDATE users
SET org_id = NULL, updated_at = now()
WHERE id = $1 AND org_id = $2;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, slug = $3
WHERE id = $1
RETURNING *;

-- name: ClearOrgUsers :exec
UPDATE users
SET org_id = NULL, updated_at = now()
WHERE org_id = $1;

-- name: DeleteRolesByOrg :exec
DELETE FROM roles WHERE org_id = $1;

-- name: DeleteQuotasByOrg :exec
DELETE FROM quotas WHERE org_id = $1;

-- name: DeleteInvitesByOrg :exec
DELETE FROM team_invites WHERE org_id = $1;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE id = $1;
