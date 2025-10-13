-- +migrate Down
-- Drop audit_logs table
DROP INDEX IF EXISTS idx_audit_logs_metadata;
DROP INDEX IF EXISTS idx_audit_logs_changes;
DROP INDEX IF EXISTS idx_audit_logs_company_created;
DROP INDEX IF EXISTS idx_audit_logs_resource_created;
DROP INDEX IF EXISTS idx_audit_logs_user_action_created;
DROP INDEX IF EXISTS idx_audit_logs_trace_id;
DROP INDEX IF EXISTS idx_audit_logs_success;
DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_resource_id;
DROP INDEX IF EXISTS idx_audit_logs_resource;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_company_id;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP TABLE IF EXISTS audit_logs;
