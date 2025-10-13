DELETE FROM users WHERE email = 'admin@test.com';

INSERT INTO users (id, name, email, password, phone, cpf, role_id, company_id, active) 
SELECT 
    gen_random_uuid(), 
    'Admin Test', 
    'admin@test.com', 
    '$2a$12$9GVupqmRFeydx4TjCwboqOL7zZQMzSL8pw0Mi.URuc4pbymow1Msi', 
    '+5511999999999', 
    '000.000.000-00', 
    r.id, 
    NULL, 
    true 
FROM roles r 
WHERE r.name = 'master';

SELECT 'User created successfully:' as status;
SELECT id, name, email, active, role_id FROM users WHERE email = 'admin@test.com';
