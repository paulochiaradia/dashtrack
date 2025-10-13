-- +migrate Up
-- Add comments to audit_logs explaining nullable fields
-- All fields are already created with correct structure in migration 006

-- Add comments explaining nullable fields
COMMENT ON COLUMN audit_logs.user_id IS 'User ID can be NULL for anonymous requests, system actions, or failed authentication attempts';
COMMENT ON COLUMN audit_logs.user_email IS 'User email can be NULL for anonymous requests';
COMMENT ON COLUMN audit_logs.company_id IS 'Company ID can be NULL for system-level actions';
