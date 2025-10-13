-- +migrate Down
-- Drop vehicles table

-- Drop indexes
DROP INDEX IF EXISTS idx_vehicles_company_availability;
DROP INDEX IF EXISTS idx_vehicles_company_status;
DROP INDEX IF EXISTS idx_vehicles_gps_device;
DROP INDEX IF EXISTS idx_vehicles_created_at;
DROP INDEX IF EXISTS idx_vehicles_vehicle_type;
DROP INDEX IF EXISTS idx_vehicles_availability;
DROP INDEX IF EXISTS idx_vehicles_status;
DROP INDEX IF EXISTS idx_vehicles_license_plate;
DROP INDEX IF EXISTS idx_vehicles_company_id;

-- Drop table
DROP TABLE IF EXISTS vehicles;
