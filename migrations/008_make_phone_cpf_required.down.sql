-- +migrate Down
-- Remove constraints and make phone/cpf optional again

-- Remove check constraints
ALTER TABLE users 
DROP CONSTRAINT IF EXISTS chk_phone_format;

ALTER TABLE users 
DROP CONSTRAINT IF EXISTS chk_cpf_format;

-- Make columns nullable again
ALTER TABLE users 
ALTER COLUMN phone DROP NOT NULL;

ALTER TABLE users 
ALTER COLUMN cpf DROP NOT NULL;
