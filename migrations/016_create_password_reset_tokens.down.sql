-- Drop password_reset_tokens table
DROP INDEX IF EXISTS idx_password_reset_valid_tokens;
DROP INDEX IF EXISTS idx_password_reset_used_at;
DROP INDEX IF EXISTS idx_password_reset_expires_at;
DROP INDEX IF EXISTS idx_password_reset_token_code;
DROP INDEX IF EXISTS idx_password_reset_user_id;

DROP TABLE IF EXISTS password_reset_tokens;
