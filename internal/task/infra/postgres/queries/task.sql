-- name: CreateTask :one
INSERT INTO tasks (title, notes, owner_id)
VALUES ($1, $2, $3)
RETURNING id, title, notes, owner_id, created_at, updated_at;

-- name: GetTask :one
SELECT id, title, notes, owner_id, created_at, updated_at
FROM tasks
WHERE id = $1 AND owner_id = $2;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2, notes = $3, updated_at = NOW()
WHERE id = $1 AND owner_id = $4
RETURNING id, title, notes, owner_id, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND owner_id = $2;

-- name: ListTasks :many
SELECT id, title, notes, owner_id, created_at, updated_at
FROM tasks
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
