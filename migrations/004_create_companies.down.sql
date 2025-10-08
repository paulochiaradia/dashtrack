-- Remove foreign key constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_company_id;

-- Drop companies table
DROP TABLE IF EXISTS companies;