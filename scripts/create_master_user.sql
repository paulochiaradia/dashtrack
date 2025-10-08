-- Script para criar usuário Master inicial
-- Execute este script no PostgreSQL após subir a aplicação

-- 1. Verificar se já existe uma empresa master
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM companies WHERE slug = 'master') THEN
        INSERT INTO companies (name, slug, email, phone, address) 
        VALUES (
            'Master Company', 
            'master', 
            'master@dashtrack.com',
            '+5511999999999',
            'Sistema DashTrack - Sede'
        );
        RAISE NOTICE 'Empresa Master criada com sucesso!';
    ELSE
        RAISE NOTICE 'Empresa Master já existe.';
    END IF;
END $$;

-- 2. Criar usuário Master inicial
DO $$
DECLARE
    master_role_id UUID;
    master_company_id UUID;
BEGIN
    -- Buscar ID do role master
    SELECT id INTO master_role_id FROM roles WHERE name = 'master' LIMIT 1;
    
    -- Buscar ID da empresa master
    SELECT id INTO master_company_id FROM companies WHERE slug = 'master' LIMIT 1;
    
    -- Verificar se o usuário master já existe
    IF NOT EXISTS (SELECT 1 FROM users WHERE email = 'master@dashtrack.com') THEN
        INSERT INTO users (
            name, 
            email, 
            password, 
            role_id, 
            company_id,
            active
        ) VALUES (
            'Master Admin',
            'master@dashtrack.com',
            '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "password"
            master_role_id,
            master_company_id,
            true
        );
        RAISE NOTICE 'Usuário Master criado com sucesso!';
        RAISE NOTICE 'Email: master@dashtrack.com';
        RAISE NOTICE 'Senha: password';
    ELSE
        RAISE NOTICE 'Usuário Master já existe.';
    END IF;
END $$;

-- 3. Verificar criação
SELECT 
    u.name,
    u.email,
    r.name as role,
    c.name as company,
    u.active,
    u.created_at
FROM users u
JOIN roles r ON u.role_id = r.id
JOIN companies c ON u.company_id = c.id
WHERE u.email = 'master@dashtrack.com';