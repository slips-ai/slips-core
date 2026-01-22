-- Remove owner_id from tags table
DROP INDEX IF EXISTS idx_tags_owner_id;
ALTER TABLE tags DROP COLUMN IF EXISTS owner_id;

-- Remove owner_id from tasks table
DROP INDEX IF EXISTS idx_tasks_owner_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS owner_id;
