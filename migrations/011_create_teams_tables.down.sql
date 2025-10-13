-- +migrate Down
-- Drop team management tables in reverse order

-- Drop indexes
DROP INDEX IF EXISTS idx_team_members_role;
DROP INDEX IF EXISTS idx_team_members_user_id;
DROP INDEX IF EXISTS idx_team_members_team_id;

DROP INDEX IF EXISTS idx_teams_name;
DROP INDEX IF EXISTS idx_teams_created_at;
DROP INDEX IF EXISTS idx_teams_status;
DROP INDEX IF EXISTS idx_teams_company_id;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;

