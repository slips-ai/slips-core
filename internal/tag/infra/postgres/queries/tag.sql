-- name: CreateTag :one
INSERT INTO tags (name)
VALUES ($1)
RETURNING id, name, created_at, updated_at;

-- name: GetTag :one
SELECT id, name, created_at, updated_at
FROM tags
WHERE id = $1;

-- name: UpdateTag :one
UPDATE tags
SET name = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, created_at, updated_at;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = $1;

-- name: ListTags :many
SELECT id, name, created_at, updated_at
FROM tags
ORDER BY name ASC
LIMIT $1 OFFSET $2;
