-- +migrate Up
-- Add manager_id column to teams table to support team manager assignment

ALTER TABLE teams ADD COLUMN IF NOT EXISTS manager_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Create index for manager lookups
CREATE INDEX IF NOT EXISTS idx_teams_manager_id ON teams(manager_id);

-- Add comment for documentation
COMMENT ON COLUMN teams.manager_id IS 'Optional team manager (user) - set to NULL if manager is deleted';
