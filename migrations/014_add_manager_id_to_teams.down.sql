-- +migrate Down
-- Remove manager_id column from teams table

DROP INDEX IF EXISTS idx_teams_manager_id;
ALTER TABLE teams DROP COLUMN IF EXISTS manager_id;
