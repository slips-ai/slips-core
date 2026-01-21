-- name: CreateTask :one
INSERT INTO tasks (title, notes)
VALUES ($1, $2)
RETURNING id, title, notes, created_at, updated_at;

-- name: GetTask :one
SELECT id, title, notes, created_at, updated_at
FROM tasks
WHERE id = $1;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2, notes = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, title, notes, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;

-- name: ListTasks :many
SELECT id, title, notes, created_at, updated_at
FROM tasks
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
