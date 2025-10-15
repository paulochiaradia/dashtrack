-- Migration: Drop team member history tracking table

DROP INDEX IF EXISTS idx_team_member_history_new_team;
DROP INDEX IF EXISTS idx_team_member_history_prev_team;
DROP INDEX IF EXISTS idx_team_member_history_change_type;
DROP INDEX IF EXISTS idx_team_member_history_changed_at;
DROP INDEX IF EXISTS idx_team_member_history_company;
DROP INDEX IF EXISTS idx_team_member_history_user;
DROP INDEX IF EXISTS idx_team_member_history_team;

DROP TABLE IF EXISTS team_member_history;
