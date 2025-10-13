-- +migrate Up
-- Create team_vehicles table for vehicle-team relationships
-- This migration runs after vehicles table is created (migration 013)

CREATE TABLE team_vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,

    -- Assignment details
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT team_vehicles_unique UNIQUE (team_id, vehicle_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_team_vehicles_team_id ON team_vehicles(team_id);
CREATE INDEX IF NOT EXISTS idx_team_vehicles_vehicle_id ON team_vehicles(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_team_vehicles_assigned_at ON team_vehicles(assigned_at DESC);

-- Add comments for documentation
COMMENT ON TABLE team_vehicles IS 'Vehicles assigned to teams for operations';
COMMENT ON COLUMN team_vehicles.assigned_by IS 'User who assigned the vehicle to the team';
