-- Migration: Drop vehicle assignment history tracking table

DROP INDEX IF EXISTS idx_vehicle_assignment_history_new_team;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_prev_team;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_new_helper;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_prev_helper;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_new_driver;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_prev_driver;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_change_type;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_changed_at;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_company;
DROP INDEX IF EXISTS idx_vehicle_assignment_history_vehicle;

DROP TABLE IF EXISTS vehicle_assignment_history;
