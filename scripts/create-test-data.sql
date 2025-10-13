-- Create test data for Team Management tests
-- This script creates users, teams, and vehicles for testing

-- Get the Master Company ID
DO $$
DECLARE
    master_company_id UUID;
    admin_user_id UUID;
    driver_role_id UUID;
    test_user_id UUID;
    test_driver_id UUID;
    test_team_id UUID;
    test_vehicle_id UUID;
BEGIN
    -- Get Master Company ID
    SELECT id INTO master_company_id FROM companies WHERE slug = 'master' LIMIT 1;
    
    -- Get Admin User ID
    SELECT id INTO admin_user_id FROM users WHERE email = 'admin@dashtrack.com' LIMIT 1;
    
    -- Get Driver Role ID
    SELECT id INTO driver_role_id FROM roles WHERE name = 'driver' LIMIT 1;
    
    -- Delete existing test users if they exist (by ID, email, or CPF)
    DELETE FROM users WHERE 
        id IN ('1b4f2ac0-d611-474d-9d25-97b3fa5369f4'::uuid, '3ac7e51e-28b0-4498-adf7-27b406b33c37'::uuid)
        OR email IN ('testuser@dashtrack.com', 'testdriver@dashtrack.com')
        OR cpf IN ('111.222.333-44', '111.222.333-45');
    
    -- Create a test user for team member operations
    INSERT INTO users (id, name, email, password, role_id, company_id, cpf, phone, active, created_at, updated_at)
    VALUES (
        '1b4f2ac0-d611-474d-9d25-97b3fa5369f4'::uuid,
        'Test User',
        'testuser@dashtrack.com',
        '$2a$12$UVS/cjIV95Lc8SIXs1o41u0T5il06vjkJ71f7GruHbXm.pqgp3Lh2', -- password: "password"
        driver_role_id,
        master_company_id,
        '111.222.333-44',
        '+5511888888888',
        true,
        NOW(),
        NOW()
    );
    
    test_user_id := '1b4f2ac0-d611-474d-9d25-97b3fa5369f4'::uuid;
    
    -- Create a test driver user
    INSERT INTO users (id, name, email, password, role_id, company_id, cpf, phone, active, created_at, updated_at)
    VALUES (
        '3ac7e51e-28b0-4498-adf7-27b406b33c37'::uuid,
        'Test Driver',
        'testdriver@dashtrack.com',
        '$2a$12$UVS/cjIV95Lc8SIXs1o41u0T5il06vjkJ71f7GruHbXm.pqgp3Lh2', -- password: "password"
        driver_role_id,
        master_company_id,
        '111.222.333-45',
        '+5511888888887',
        true,
        NOW(),
        NOW()
    );
    
    test_driver_id := '3ac7e51e-28b0-4498-adf7-27b406b33c37'::uuid;
    
    -- Delete existing test vehicle if it exists
    DELETE FROM vehicles WHERE id = '9c6ded57-61df-4fb5-97f0-a25d2898fc89'::uuid;
    
    -- Create a test vehicle
    INSERT INTO vehicles (
        id, company_id, license_plate, brand, model, year, color,
        vehicle_type, fuel_type, cargo_capacity, status,
        created_at, updated_at
    )
    VALUES (
        '9c6ded57-61df-4fb5-97f0-a25d2898fc89'::uuid,
        master_company_id,
        'ABC-1234',
        'Ford',
        'F-150',
        2023,
        'White',
        'truck',
        'diesel',
        1000.50,
        'active',
        NOW(),
        NOW()
    );
    
    test_vehicle_id := '9c6ded57-61df-4fb5-97f0-a25d2898fc89'::uuid;
    
    RAISE NOTICE 'Test data created successfully:';
    RAISE NOTICE '  Master Company ID: %', master_company_id;
    RAISE NOTICE '  Admin User ID: %', admin_user_id;
    RAISE NOTICE '  Test User ID: %', test_user_id;
    RAISE NOTICE '  Test Driver ID: %', test_driver_id;
    RAISE NOTICE '  Test Vehicle ID: %', test_vehicle_id;
END $$;
