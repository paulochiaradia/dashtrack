-- +migrate Up
-- Create auth_logs table
CREATE TABLE IF NOT EXISTS auth_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL,
    email_attempt VARCHAR(100) NOT NULL,
    success BOOLEAN NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    failure_reason VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for auth_logs table
CREATE INDEX IF NOT EXISTS idx_auth_logs_user_created ON auth_logs(user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_auth_logs_email_created ON auth_logs(email_attempt, created_at);
CREATE INDEX IF NOT EXISTS idx_auth_logs_success_created ON auth_logs(success, created_at);

-- +migrate Down
DROP INDEX IF EXISTS idx_auth_logs_success_created;
DROP INDEX IF EXISTS idx_auth_logs_email_created;
DROP INDEX IF EXISTS idx_auth_logs_user_created;
DROP TABLE IF EXISTS auth_logs;
