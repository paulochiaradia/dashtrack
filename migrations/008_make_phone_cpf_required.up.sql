-- +migrate Up
-- Make phone and cpf fields required for all users

-- First, update existing users that have NULL or empty phone/cpf
-- This is needed before adding NOT NULL constraints

-- Update users with NULL phone - set placeholder with different DDDs
UPDATE users 
SET phone = CASE 
    WHEN name = 'Master Admin' THEN '+5511000000001'    -- São Paulo
    WHEN name = 'General Admin' THEN '+5521000000002'   -- Rio de Janeiro
    WHEN email LIKE '%admin%' THEN '+5547000000003'     -- Santa Catarina
    WHEN email LIKE '%manager%' THEN '+5531000000004'   -- Minas Gerais
    ELSE '+55' || LPAD((RANDOM() * 89 + 11)::INT::TEXT, 2, '0') || '000000000'  -- Random DDD
END
WHERE phone IS NULL OR phone = '';

-- Update users with NULL cpf - set placeholder CPFs
UPDATE users 
SET cpf = CASE 
    WHEN name = 'Master Admin' THEN '000.000.000-00'
    WHEN name = 'General Admin' THEN '111.111.111-11'
    ELSE '999.999.999-99'
END
WHERE cpf IS NULL OR cpf = '';

-- Now add NOT NULL constraints
ALTER TABLE users 
ALTER COLUMN phone SET NOT NULL;

ALTER TABLE users 
ALTER COLUMN cpf SET NOT NULL;

-- Add check constraints for phone format (basic validation)
ALTER TABLE users 
ADD CONSTRAINT chk_phone_format 
CHECK (phone ~ '^\+[1-9]\d{1,14}$');

-- Add check constraint for CPF format (Brazilian format)
ALTER TABLE users 
ADD CONSTRAINT chk_cpf_format 
CHECK (cpf ~ '^\d{3}\.\d{3}\.\d{3}-\d{2}$');
