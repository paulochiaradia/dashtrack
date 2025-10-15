-- Migration: Create vehicle assignment history tracking table
-- This table logs all changes to vehicle assignments (driver, helper, team)

CREATE TABLE IF NOT EXISTS vehicle_assignment_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Previous assignments
    previous_driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    previous_helper_id UUID REFERENCES users(id) ON DELETE SET NULL,
    previous_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    
    -- New assignments
    new_driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    new_helper_id UUID REFERENCES users(id) ON DELETE SET NULL,
    new_team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    
    -- Change metadata
    change_type VARCHAR(50) NOT NULL CHECK (change_type IN ('driver', 'helper', 'team', 'full_assignment')),
    changed_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    change_reason TEXT,
    changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Audit fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_vehicle ON vehicle_assignment_history(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_company ON vehicle_assignment_history(company_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_changed_at ON vehicle_assignment_history(changed_at DESC);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_change_type ON vehicle_assignment_history(change_type);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_prev_driver ON vehicle_assignment_history(previous_driver_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_new_driver ON vehicle_assignment_history(new_driver_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_prev_helper ON vehicle_assignment_history(previous_helper_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_new_helper ON vehicle_assignment_history(new_helper_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_prev_team ON vehicle_assignment_history(previous_team_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_assignment_history_new_team ON vehicle_assignment_history(new_team_id);

-- Comments for documentation
COMMENT ON TABLE vehicle_assignment_history IS 'Tracks historical changes to vehicle assignments';
COMMENT ON COLUMN vehicle_assignment_history.change_type IS 'Type of change: driver (driver only), helper (helper only), team (team only), full_assignment (multiple fields)';
COMMENT ON COLUMN vehicle_assignment_history.changed_by_user_id IS 'User who made the change (usually admin or company admin)';
COMMENT ON COLUMN vehicle_assignment_history.change_reason IS 'Optional reason for the assignment change';
