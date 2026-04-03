-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = now(), updated_at = now()
WHERE id = $1;

-- name: UpdateUserMFA :exec
UPDATE users
SET mfa_enabled = $2,
    mfa_secret_encrypted = $3,
    updated_at = now()
WHERE id = $1;

-- name: ListUserRoles :many
SELECT r.*
FROM roles r
JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1
ORDER BY r.name;

-- name: ListUsers :many
SELECT
  u.id,
  u.org_id,
  u.email,
  u.full_name,
  u.mfa_enabled,
  u.status,
  u.created_at,
  u.updated_at,
  u.last_login_at,
  COALESCE(array_remove(array_agg(r.name), NULL), ARRAY[]::text[])::text[] AS role_names
FROM users u
LEFT JOIN user_roles ur ON ur.user_id = u.id
LEFT JOIN roles r ON r.id = ur.role_id
WHERE u.org_id = $1
GROUP BY u.id
ORDER BY u.created_at DESC;

-- name: ListRoles :many
SELECT * FROM roles
WHERE org_id = $1
ORDER BY name;

-- name: DeleteUserRolesByUserID :exec
DELETE FROM user_roles
WHERE user_id = $1;

-- name: AddUserRole :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: ListPermissionsByRoleIDs :many
SELECT DISTINCT p.*
FROM permissions p
JOIN role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = ANY($1::uuid[])
ORDER BY p.key;

-- name: IncrementFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = COALESCE(failed_login_attempts, 0) + 1,
    updated_at = now()
WHERE id = $1;

-- name: ResetFailedLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0,
    locked_until = NULL,
    updated_at = now()
WHERE id = $1;

-- name: LockUserAccount :exec
UPDATE users
SET locked_until = $2,
    updated_at = now()
WHERE id = $1;
