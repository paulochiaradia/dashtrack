-- +migrate Down
-- Drop team_vehicles table

DROP INDEX IF EXISTS idx_team_vehicles_assigned_at;
DROP INDEX IF EXISTS idx_team_vehicles_vehicle_id;
DROP INDEX IF EXISTS idx_team_vehicles_team_id;

DROP TABLE IF EXISTS team_vehicles;
