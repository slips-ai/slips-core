DROP INDEX IF EXISTS idx_tasks_start_date_kind;
ALTER TABLE tasks DROP COLUMN IF EXISTS start_date_kind;
