-- +migrate Up
-- Add assignment columns to vehicles table

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id) ON DELETE SET NULL;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS driver_id UUID REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS helper_id UUID REFERENCES users(id) ON DELETE SET NULL;

-- Create indexes for assignment queries
CREATE INDEX IF NOT EXISTS idx_vehicles_team_id ON vehicles(team_id) WHERE team_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_driver_id ON vehicles(driver_id) WHERE driver_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_helper_id ON vehicles(helper_id) WHERE helper_id IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN vehicles.team_id IS 'Team to which the vehicle is assigned';
COMMENT ON COLUMN vehicles.driver_id IS 'Primary driver assigned to the vehicle';
COMMENT ON COLUMN vehicles.helper_id IS 'Helper/assistant assigned to the vehicle';
