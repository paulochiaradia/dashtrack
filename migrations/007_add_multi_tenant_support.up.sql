-- +migrate Up
-- +migrate Up
-- Add company_id column to users table for multi-tenant support
ALTER TABLE users ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id) ON DELETE SET NULL;

-- Create index for company_id
CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);

-- Note: Default company creation removed - create companies manually as needed
-- This allows for a completely clean database on first run
