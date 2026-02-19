-- name: CreateTask :one
INSERT INTO tasks (title, notes, owner_id, start_date)
VALUES ($1, $2, $3, $4)
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at, start_date;

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
SELECT id, title, notes, owner_id, archived_at, created_at, updated_at, start_date
FROM tasks
WHERE id = $1 AND owner_id = $2;

-- name: UpdateTask :one
UPDATE tasks
SET title = $2, notes = $3, updated_at = NOW(), start_date = $5
WHERE id = $1 AND owner_id = $4
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at, start_date;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND owner_id = $2;

-- name: ListTasks :many
SELECT DISTINCT t.id, t.title, t.notes, t.owner_id, t.archived_at, t.created_at, t.updated_at, t.start_date
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
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at, start_date;

-- name: UnarchiveTask :one
UPDATE tasks
SET archived_at = NULL, updated_at = NOW()
WHERE id = $1 AND owner_id = $2
RETURNING id, title, notes, owner_id, archived_at, created_at, updated_at, start_date;

-- name: ListChecklistItems :many
SELECT ci.*
FROM task_checklist_items ci
JOIN tasks t ON ci.task_id = t.id
WHERE ci.task_id = sqlc.arg(task_id) AND t.owner_id = sqlc.arg(owner_id)
ORDER BY ci.sort_order ASC, ci.created_at ASC;

-- name: AddChecklistItem :one
INSERT INTO task_checklist_items (task_id, content, completed, sort_order)
SELECT sqlc.arg(task_id), sqlc.arg(content), FALSE,
       COALESCE((SELECT MAX(sort_order) + 1 FROM task_checklist_items WHERE task_id = sqlc.arg(task_id)), 0)
FROM tasks
WHERE id = sqlc.arg(task_id) AND owner_id = sqlc.arg(owner_id)
RETURNING *;

-- name: CreateChecklistItemWithSortOrder :one
INSERT INTO task_checklist_items (task_id, content, completed, sort_order)
SELECT sqlc.arg(task_id), sqlc.arg(content), FALSE, sqlc.arg(sort_order)
FROM tasks
WHERE id = sqlc.arg(task_id) AND owner_id = sqlc.arg(owner_id)
RETURNING *;

-- name: UpdateChecklistItemContent :one
UPDATE task_checklist_items ci
SET content = sqlc.arg(content), updated_at = NOW()
FROM tasks t
WHERE ci.id = sqlc.arg(item_id)
  AND ci.task_id = t.id
  AND t.owner_id = sqlc.arg(owner_id)
RETURNING ci.*;

-- name: SetChecklistItemCompleted :one
UPDATE task_checklist_items ci
SET completed = sqlc.arg(completed), updated_at = NOW()
FROM tasks t
WHERE ci.id = sqlc.arg(item_id)
  AND ci.task_id = t.id
  AND t.owner_id = sqlc.arg(owner_id)
RETURNING ci.*;

-- name: DeleteChecklistItem :execrows
DELETE FROM task_checklist_items ci
USING tasks t
WHERE ci.id = sqlc.arg(item_id)
  AND ci.task_id = t.id
  AND t.owner_id = sqlc.arg(owner_id);

-- name: ReorderChecklistItems :exec
UPDATE task_checklist_items ci
SET sort_order = (ordered.ord - 1)::int,
    updated_at = NOW()
FROM unnest(sqlc.arg(item_ids)::uuid[]) WITH ORDINALITY AS ordered(id, ord)
JOIN tasks t ON t.id = sqlc.arg(task_id) AND t.owner_id = sqlc.arg(owner_id)
WHERE ci.task_id = sqlc.arg(task_id)
  AND ci.id = ordered.id;
