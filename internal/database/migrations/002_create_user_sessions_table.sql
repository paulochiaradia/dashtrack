-- +migrate Up
-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(128) PRIMARY KEY,
    user_id UUID NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    session_data JSONB,
    expires_at TIMESTAMPTZ NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for user_sessions table
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_active ON user_sessions(user_id, active);

-- +migrate Down
DROP INDEX IF EXISTS idx_user_sessions_user_active;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP TABLE IF EXISTS user_sessions;
