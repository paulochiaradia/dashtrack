-- Migration: Create team member history tracking table
-- This table logs all changes to team memberships (additions, removals, role changes)

CREATE TABLE IF NOT EXISTS team_member_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Previous state
    previous_role_in_team VARCHAR(50),
    
    -- New state
    new_role_in_team VARCHAR(50),
    
    -- Change metadata
    change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('added', 'removed', 'role_changed', 'transferred_in', 'transferred_out')),
    previous_team_id UUID REFERENCES teams(id) ON DELETE SET NULL, -- Only for transfers
    new_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,      -- Only for transfers
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    change_reason TEXT,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_team_member_history_team ON team_member_history(team_id);
CREATE INDEX IF NOT EXISTS idx_team_member_history_user ON team_member_history(user_id);
CREATE INDEX IF NOT EXISTS idx_team_member_history_company ON team_member_history(company_id);
CREATE INDEX IF NOT EXISTS idx_team_member_history_changed_at ON team_member_history(changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_team_member_history_change_type ON team_member_history(change_type);
CREATE INDEX IF NOT EXISTS idx_team_member_history_prev_team ON team_member_history(previous_team_id);
CREATE INDEX IF NOT EXISTS idx_team_member_history_new_team ON team_member_history(new_team_id);

-- Comments for documentation
COMMENT ON TABLE team_member_history IS 'Tracks historical changes to team memberships';
COMMENT ON COLUMN team_member_history.change_type IS 'Type of change: added (new member), removed (member left), role_changed (role updated), transferred_in (from another team), transferred_out (to another team)';
COMMENT ON COLUMN team_member_history.previous_team_id IS 'Source team for transfers (only populated for transferred_in)';
COMMENT ON COLUMN team_member_history.new_team_id IS 'Destination team for transfers (only populated for transferred_out)';
COMMENT ON COLUMN team_member_history.changed_by_user_id IS 'User who made the change (usually admin or company admin)';
