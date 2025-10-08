-- +migrate Down
-- +migrate Down
-- Remove foreign key constraint and company_id column
ALTER TABLE users DROP COLUMN IF EXISTS company_id;
