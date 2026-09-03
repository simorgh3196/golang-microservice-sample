-- name: GetApiKeyByHash :one
SELECT
    id,
    tenant_id,
    key_hash,
    name,
    role,
    is_active,
    created_at,
    expires_at
FROM
    api_keys
WHERE
    key_hash = $1
    AND is_active = true
LIMIT 1;

-- name: GetTenantByID :one
SELECT
    id,
    name,
    plan,
    created_at,
    updated_at
FROM
    tenants
WHERE
    id = $1
LIMIT 1;

