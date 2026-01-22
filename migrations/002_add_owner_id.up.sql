-- Add owner_id to tasks table
ALTER TABLE tasks ADD COLUMN owner_id VARCHAR(255) NOT NULL DEFAULT '';

-- Create index on owner_id for performance
CREATE INDEX IF NOT EXISTS idx_tasks_owner_id ON tasks(owner_id);

-- Add owner_id to tags table
ALTER TABLE tags ADD COLUMN owner_id VARCHAR(255) NOT NULL DEFAULT '';

-- Create index on owner_id for performance
CREATE INDEX IF NOT EXISTS idx_tags_owner_id ON tags(owner_id);
