-- +migrate Down
-- +migrate Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS session_tokens;
DROP TABLE IF EXISTS rate_limit_events;
DROP TABLE IF EXISTS rate_limit_rules;
DROP TABLE IF EXISTS two_factor_auth;
