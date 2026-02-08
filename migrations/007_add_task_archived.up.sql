-- Add archived_at column to tasks table
ALTER TABLE tasks ADD COLUMN archived_at TIMESTAMPTZ;

-- Create index for efficient filtering by archived status
CREATE INDEX IF NOT EXISTS idx_tasks_archived_at ON tasks(archived_at);

-- Create compound index for owner_id + archived_at for common queries
CREATE INDEX IF NOT EXISTS idx_tasks_owner_archived ON tasks(owner_id, archived_at);
