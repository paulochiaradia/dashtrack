-- +migrate Up
-- Create teams table for team management
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'archived')),
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT teams_company_name_unique UNIQUE (company_id, name, deleted_at)
);

-- Create team_members table for user-team relationships
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_in_team VARCHAR(50) NOT NULL CHECK (role_in_team IN ('manager', 'driver', 'assistant', 'supervisor')),
    
    -- Metadata
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT team_members_unique UNIQUE (team_id, user_id)
);

-- Note: team_vehicles table will be created in migration 015 after vehicles table exists-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_teams_company_id ON teams(company_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teams_status ON teams(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teams_created_at ON teams(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members(role_in_team);

-- Add comments for documentation
COMMENT ON TABLE teams IS 'Teams for organizing users and vehicles within a company';
COMMENT ON TABLE team_members IS 'Users assigned to teams with specific roles';

COMMENT ON COLUMN teams.status IS 'Team status: active (operational), inactive (temporarily disabled), archived (historical record)';
COMMENT ON COLUMN team_members.role_in_team IS 'Role within the team: manager (team leader), driver (vehicle operator), assistant (support), supervisor (overseer)';
