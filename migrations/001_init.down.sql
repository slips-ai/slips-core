-- Drop tables in reverse order
DROP INDEX IF EXISTS idx_tags_name;
DROP TABLE IF EXISTS tags;

DROP INDEX IF EXISTS idx_tasks_created_at;
DROP TABLE IF EXISTS tasks;
