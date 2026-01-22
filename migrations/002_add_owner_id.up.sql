-- Add owner_id to tasks table
-- NOTE: This migration sets owner_id to a sentinel value ('MIGRATED_NO_OWNER') for existing records.
-- Before running this migration in production with existing data, you should either:
-- 1. Delete all existing data, OR
-- 2. Manually assign owner_id values to existing records after migration
ALTER TABLE tasks ADD COLUMN owner_id VARCHAR(255) NOT NULL DEFAULT 'MIGRATED_NO_OWNER';

-- Create index on owner_id for performance
CREATE INDEX IF NOT EXISTS idx_tasks_owner_id ON tasks(owner_id);

-- Add owner_id to tags table
-- NOTE: Same consideration as above applies to existing tag records
ALTER TABLE tags ADD COLUMN owner_id VARCHAR(255) NOT NULL DEFAULT 'MIGRATED_NO_OWNER';

-- Create index on owner_id for performance
CREATE INDEX IF NOT EXISTS idx_tags_owner_id ON tags(owner_id);
