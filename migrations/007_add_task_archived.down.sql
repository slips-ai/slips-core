-- Remove indexes
DROP INDEX IF EXISTS idx_tasks_owner_archived;
DROP INDEX IF EXISTS idx_tasks_archived_at;

-- Remove archived_at column
ALTER TABLE tasks DROP COLUMN archived_at;
