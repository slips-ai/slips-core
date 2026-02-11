-- Add start_date_kind column (inbox or specific_date)
ALTER TABLE tasks ADD COLUMN start_date_kind VARCHAR(20) NOT NULL DEFAULT 'inbox';

-- Add start_date column (date only, no timezone)
ALTER TABLE tasks ADD COLUMN start_date DATE;

-- Create index for filtering by start_date_kind
CREATE INDEX idx_tasks_start_date_kind ON tasks(start_date_kind);

-- Create index for filtering/sorting by start_date
CREATE INDEX idx_tasks_start_date ON tasks(start_date) WHERE start_date IS NOT NULL;
