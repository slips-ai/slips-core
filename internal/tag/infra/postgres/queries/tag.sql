-- name: CreateTag :one
INSERT INTO tags (name, owner_id)
VALUES ($1, $2)
RETURNING id, name, owner_id, created_at, updated_at;

-- name: GetTag :one
SELECT id, name, owner_id, created_at, updated_at
FROM tags
WHERE id = $1 AND owner_id = $2;

-- name: UpdateTag :one
UPDATE tags
SET name = $2, updated_at = NOW()
WHERE id = $1 AND owner_id = $3
RETURNING id, name, owner_id, created_at, updated_at;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1 AND owner_id = $2;

-- name: ListTags :many
SELECT id, name, owner_id, created_at, updated_at
FROM tags
WHERE owner_id = $1
ORDER BY name ASC
LIMIT $2 OFFSET $3;
