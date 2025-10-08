-- +migrate Up
-- +migrate Up
-- Add company_id column to users table for multi-tenant support
ALTER TABLE users ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id) ON DELETE SET NULL;

-- Create index for company_id
CREATE INDEX IF NOT EXISTS idx_users_company_id ON users(company_id);

-- Insert a default company if not exists
INSERT INTO companies (name, slug, email) 
SELECT 'Default Company', 'default', 'admin@defaultcompany.com'
WHERE NOT EXISTS (SELECT 1 FROM companies WHERE slug = 'default');

-- Update existing users to have the default company
UPDATE users SET company_id = (SELECT id FROM companies WHERE slug = 'default') 
WHERE company_id IS NULL;
