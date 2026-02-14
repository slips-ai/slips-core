ALTER TABLE tasks ADD COLUMN start_date_kind VARCHAR(20) NOT NULL DEFAULT 'inbox';
UPDATE tasks
SET start_date_kind = CASE
    WHEN start_date IS NULL THEN 'inbox'
    ELSE 'specific_date'
END;
CREATE INDEX IF NOT EXISTS idx_tasks_start_date_kind ON tasks(start_date_kind);
