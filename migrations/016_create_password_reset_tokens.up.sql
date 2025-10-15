-- Create password_reset_tokens table for email-based password recovery
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_code VARCHAR(6) NOT NULL, -- Código de 6 dígitos
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip_address VARCHAR(45), -- Suporte IPv4 e IPv6
    user_agent TEXT,
    
    -- Constraints
    CONSTRAINT chk_token_code_format CHECK (token_code ~ '^[0-9]{6}$'),
    CONSTRAINT chk_expires_after_creation CHECK (expires_at > created_at),
    CONSTRAINT chk_used_after_creation CHECK (used_at IS NULL OR used_at >= created_at)
);

-- Índices para performance
CREATE INDEX idx_password_reset_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_token_code ON password_reset_tokens(token_code);
CREATE INDEX idx_password_reset_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX idx_password_reset_used_at ON password_reset_tokens(used_at);

-- Índice composto para busca de tokens válidos
CREATE INDEX idx_password_reset_valid_tokens ON password_reset_tokens(user_id, token_code, expires_at, used_at);

-- Comentários
COMMENT ON TABLE password_reset_tokens IS 'Tokens de recuperação de senha enviados por email';
COMMENT ON COLUMN password_reset_tokens.token_code IS 'Código de 6 dígitos enviado por email';
COMMENT ON COLUMN password_reset_tokens.expires_at IS 'Data/hora de expiração (15 minutos após criação)';
COMMENT ON COLUMN password_reset_tokens.used_at IS 'Data/hora em que o token foi usado (NULL = não usado)';
COMMENT ON COLUMN password_reset_tokens.ip_address IS 'Endereço IP de onde a solicitação foi feita';
COMMENT ON COLUMN password_reset_tokens.user_agent IS 'User agent do navegador/cliente';
