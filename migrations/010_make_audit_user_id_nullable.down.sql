-- +migrate Down
-- Revert user_id back to NOT NULL (be careful - this might fail if NULL values exist)
-- First, update NULL values to a default user or delete them
DELETE FROM audit_logs WHERE user_id IS NULL;

-- Then restore NOT NULL constraint
ALTER TABLE audit_logs ALTER COLUMN user_id SET NOT NULL;
