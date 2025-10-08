-- DashTrack - Setup Inicial do Sistema
-- Execute este script após criar o banco limpo para configurar o sistema

-- 1. Verificar se as tabelas existem
\dt

-- 2. Verificar roles disponíveis
SELECT * FROM roles;

-- 3. Criar empresa master (ajuste os dados conforme necessário)
INSERT INTO companies (name, slug, email, phone, address, subscription_plan, max_users, max_vehicles, max_sensors) 
VALUES (
    'Master Company', 
    'master', 
    'master@dashtrack.com',
    '+5511999999999',
    'Sede da Empresa, São Paulo, SP',
    'enterprise', 
    1000, 
    1000, 
    1000
) ON CONFLICT (slug) DO NOTHING;

-- 4. Criar usuário master inicial
INSERT INTO users (name, email, password, phone, role_id, company_id, active) 
SELECT 
    'Master Admin',
    'master@dashtrack.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: "password"
    '+5511888888888',
    r.id,
    c.id,
    true
FROM roles r, companies c 
WHERE r.name = 'master' AND c.slug = 'master'
ON CONFLICT (email) DO NOTHING;

-- 5. Verificar se foi criado corretamente
SELECT u.name, u.email, c.name as company, r.name as role
FROM users u 
JOIN companies c ON u.company_id = c.id 
JOIN roles r ON u.role_id = r.id
WHERE u.email = 'master@dashtrack.com';

-- 6. Listar empresas criadas
SELECT id, name, slug, email, subscription_plan, max_users FROM companies;

\echo '✅ Setup inicial concluído!'
\echo '📧 Email: master@dashtrack.com'
\echo '🔑 Senha: password'
\echo '🌐 URL: http://localhost:8080/api/v1/auth/login'