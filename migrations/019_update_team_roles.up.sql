-- Update team_members role_in_team constraint to include helper and team_lead

-- Drop the old constraint
ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_role_in_team_check;

-- Add new constraint with all valid roles
ALTER TABLE team_members ADD CONSTRAINT team_members_role_in_team_check 
    CHECK (role_in_team IN ('manager', 'driver', 'assistant', 'supervisor', 'helper', 'team_lead'));

-- Update comment
COMMENT ON COLUMN team_members.role_in_team IS 'Role within the team: manager (team manager), driver (vehicle operator), assistant (support), supervisor (overseer), helper (assistant), team_lead (team leader)';
