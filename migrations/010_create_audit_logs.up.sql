-- +migrate Up
-- Create audit_logs table for comprehensive system auditing
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_email VARCHAR(255) NOT NULL,
    company_id UUID REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Action details
    action VARCHAR(50) NOT NULL, -- CREATE, UPDATE, DELETE, READ, LOGIN, LOGOUT, etc
    resource VARCHAR(100) NOT NULL, -- user, vehicle, team, company, etc
    resource_id UUID,
    
    -- Request context
    method VARCHAR(10) NOT NULL, -- GET, POST, PUT, DELETE, PATCH
    path VARCHAR(500) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    
    -- Data changes
    changes JSONB, -- Store before/after state
    metadata JSONB, -- Additional context
    
    -- Result
    success BOOLEAN NOT NULL DEFAULT true,
    error_message TEXT,
    status_code INTEGER NOT NULL,
    duration_ms BIGINT, -- Duration in milliseconds
    
    -- Distributed tracing
    trace_id VARCHAR(32), -- Jaeger trace ID
    span_id VARCHAR(16), -- Jaeger span ID
    
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_company_id ON audit_logs(company_id) WHERE company_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id) WHERE resource_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_success ON audit_logs(success);
CREATE INDEX IF NOT EXISTS idx_audit_logs_trace_id ON audit_logs(trace_id) WHERE trace_id IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_action_created ON audit_logs(user_id, action, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_created ON audit_logs(resource, resource_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_company_created ON audit_logs(company_id, created_at DESC) WHERE company_id IS NOT NULL;

-- GIN index for JSON queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_changes ON audit_logs USING GIN(changes) WHERE changes IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_logs_metadata ON audit_logs USING GIN(metadata) WHERE metadata IS NOT NULL;

-- Comments
COMMENT ON TABLE audit_logs IS 'Comprehensive audit log for all user actions and system events';
COMMENT ON COLUMN audit_logs.action IS 'Type of action: CREATE, UPDATE, DELETE, READ, LOGIN, LOGOUT, etc';
COMMENT ON COLUMN audit_logs.resource IS 'Resource type affected: user, vehicle, team, company, etc';
COMMENT ON COLUMN audit_logs.changes IS 'JSON containing before/after state for UPDATE operations';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional context data in JSON format';
COMMENT ON COLUMN audit_logs.trace_id IS 'Jaeger distributed tracing trace ID for correlation';
COMMENT ON COLUMN audit_logs.duration_ms IS 'Operation duration in milliseconds';
