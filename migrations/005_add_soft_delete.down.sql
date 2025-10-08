-- Remove soft delete support
DROP INDEX IF EXISTS idx_companies_deleted_at;
ALTER TABLE companies DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;