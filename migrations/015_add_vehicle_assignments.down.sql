-- +migrate Down
-- Remove assignment columns from vehicles table

DROP INDEX IF EXISTS idx_vehicles_team_id;
DROP INDEX IF EXISTS idx_vehicles_driver_id;
DROP INDEX IF EXISTS idx_vehicles_helper_id;

ALTER TABLE vehicles DROP COLUMN IF EXISTS team_id;
ALTER TABLE vehicles DROP COLUMN IF EXISTS driver_id;
ALTER TABLE vehicles DROP COLUMN IF EXISTS helper_id;
