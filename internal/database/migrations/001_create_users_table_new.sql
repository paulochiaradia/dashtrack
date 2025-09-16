-- +migrate Up
-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default roles
INSERT INTO roles (name, description) VALUES 
    ('admin', 'System administrator with full access'),
    ('driver', 'Vehicle driver with access to assigned vehicles'),
    ('helper', 'Driver assistant with limited access to assigned vehicles')
ON CONFLICT (name) DO NOTHING;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    cpf VARCHAR(14) UNIQUE,
    avatar VARCHAR(255),
    role_id UUID NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMPTZ NULL,
    dashboard_config JSONB,
    api_token VARCHAR(255) UNIQUE,
    login_attempts INT DEFAULT 0,
    blocked_until TIMESTAMPTZ NULL,
    password_changed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(active);
CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

-- +migrate Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
