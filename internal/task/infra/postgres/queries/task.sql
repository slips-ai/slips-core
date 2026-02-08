-- name: CreateTask :one
INSERT INTO tasks (title, notes, owner_id)
VALUES ($1, $2, $3)
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at;

-- name: CreateTaskTag :exec
INSERT INTO task_tags (task_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteTaskTags :exec
DELETE FROM task_tags
WHERE task_id = $1;

-- name: GetTaskTagIDs :many
SELECT tag_id
FROM task_tags
WHERE task_id = $1;

-- name: GetTask :one
SELECT id, title, notes, owner_id, archived_at, created_at, updated_at
FROM tasks
WHERE id = $1 AND owner_id = $2;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2, notes = $3, updated_at = NOW()
WHERE id = $1 AND owner_id = $4
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND owner_id = $2;

-- name: ListTasks :many
SELECT DISTINCT t.id, t.title, t.notes, t.owner_id, t.archived_at, t.created_at, t.updated_at
FROM tasks t
LEFT JOIN task_tags tt ON t.id = tt.task_id
WHERE t.owner_id = $1
  AND (sqlc.narg('filter_tag_ids')::uuid[] IS NULL
       OR tt.tag_id = ANY(sqlc.narg('filter_tag_ids')::uuid[]))
  AND (
    (sqlc.narg('archived_only')::boolean = TRUE AND t.archived_at IS NOT NULL) OR
    (sqlc.narg('archived_only')::boolean = FALSE AND (
      sqlc.narg('include_archived')::boolean = TRUE OR
      (sqlc.narg('include_archived')::boolean = FALSE AND t.archived_at IS NULL)
    )) OR
    (sqlc.narg('archived_only')::boolean IS NULL AND sqlc.narg('include_archived')::boolean IS NULL AND t.archived_at IS NULL)
  )
ORDER BY t.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ArchiveTask :one
UPDATE tasks
SET archived_at = NOW(), updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at;

-- name: UnarchiveTask :one
UPDATE tasks
SET archived_at = NULL, updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at;
