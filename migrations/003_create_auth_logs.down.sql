-- +migrate Down
DROP INDEX IF EXISTS idx_auth_logs_success_created;
DROP INDEX IF EXISTS idx_auth_logs_email_created;
DROP INDEX IF EXISTS idx_auth_logs_user_created;
DROP TABLE IF EXISTS auth_logs;
