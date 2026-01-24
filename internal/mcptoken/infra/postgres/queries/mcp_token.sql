-- name: CreateMCPToken :one
INSERT INTO mcp_tokens (token, user_id, name, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, token, user_id, name, created_at, expires_at, last_used_at, is_active;

-- name: GetMCPTokenByToken :one
SELECT id, token, user_id, name, created_at, expires_at, last_used_at, is_active
FROM mcp_tokens
WHERE token = $1;

-- name: GetMCPTokenByID :one
SELECT id, token, user_id, name, created_at, expires_at, last_used_at, is_active
FROM mcp_tokens
WHERE id = $1;

-- name: ListMCPTokensByUserID :many
SELECT id, token, user_id, name, created_at, expires_at, last_used_at, is_active
FROM mcp_tokens
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateMCPTokenLastUsedAt :exec
UPDATE mcp_tokens
SET last_used_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: RevokeMCPToken :exec
UPDATE mcp_tokens
SET is_active = FALSE
WHERE id = $1;

-- name: DeleteMCPToken :exec
DELETE FROM mcp_tokens
WHERE id = $1;
