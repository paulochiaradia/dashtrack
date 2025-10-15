-- Rollback team_members role_in_team constraint update

-- Drop the new constraint
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_role_in_team_check;

-- Restore original constraint
ALTER TABLE team_members ADD CONSTRAINT team_members_role_in_team_check 
    CHECK (role_in_team IN ('manager', 'driver', 'assistant', 'supervisor'));

-- Restore original comment
COMMENT ON COLUMN team_members.role_in_team IS 'Role within the team: manager (team leader), driver (vehicle operator), assistant (support), supervisor (overseer)';
