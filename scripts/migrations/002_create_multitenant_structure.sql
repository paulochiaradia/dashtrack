-- Migration: Create Multi-Tenant Company Structure
-- Date: 2025-09-29
-- Description: Create company hierarchy and multi-tenant support

-- Create companies table
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL, -- URL-friendly identifier
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(50),
    country VARCHAR(50) DEFAULT 'Brazil',
    subscription_plan VARCHAR(50) DEFAULT 'basic' CHECK (subscription_plan IN ('basic', 'premium', 'enterprise')),
    max_users INTEGER DEFAULT 50,
    max_vehicles INTEGER DEFAULT 20,
    max_sensors INTEGER DEFAULT 100,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for companies
CREATE INDEX IF NOT EXISTS idx_companies_slug ON companies(slug);
CREATE INDEX IF NOT EXISTS idx_companies_status ON companies(status);
CREATE INDEX IF NOT EXISTS idx_companies_plan ON companies(subscription_plan);

-- Add company_id to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id) ON DELETE CASCADE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS employee_id VARCHAR(50); -- Internal company employee ID
ALTER TABLE users ADD COLUMN IF NOT EXISTS department VARCHAR(100); -- HR department info

-- Update user roles to include company context
ALTER TABLE users DROP CONSTRAINT IF EXISTS check_role;
ALTER TABLE users ADD CONSTRAINT check_role CHECK (role IN ('master', 'company_admin', 'manager', 'driver', 'helper', 'user'));

-- Create teams table
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    manager_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for teams
CREATE INDEX IF NOT EXISTS idx_teams_company_id ON teams(company_id);
CREATE INDEX IF NOT EXISTS idx_teams_manager_id ON teams(manager_id);
CREATE INDEX IF NOT EXISTS idx_teams_status ON teams(status);

-- Create team_members table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS team_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_in_team VARCHAR(50) DEFAULT 'member' CHECK (role_in_team IN ('leader', 'driver', 'helper', 'member')),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, user_id)
);

-- Create indexes for team_members
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);
CREATE INDEX IF NOT EXISTS idx_team_members_role ON team_members(role_in_team);

-- Create vehicles table
CREATE TABLE IF NOT EXISTS vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    license_plate VARCHAR(20) UNIQUE NOT NULL,
    brand VARCHAR(100),
    model VARCHAR(100),
    year INTEGER,
    color VARCHAR(50),
    vehicle_type VARCHAR(50) DEFAULT 'truck' CHECK (vehicle_type IN ('truck', 'van', 'car', 'motorcycle', 'bus')),
    fuel_type VARCHAR(30) DEFAULT 'diesel' CHECK (fuel_type IN ('gasoline', 'diesel', 'electric', 'hybrid', 'cng')),
    capacity_kg DECIMAL(10,2), -- Load capacity in kg
    driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    helper_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'maintenance', 'inactive', 'retired')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for vehicles
CREATE INDEX IF NOT EXISTS idx_vehicles_company_id ON vehicles(company_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_team_id ON vehicles(team_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_driver_id ON vehicles(driver_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_helper_id ON vehicles(helper_id);
CREATE INDEX IF NOT EXISTS idx_vehicles_plate ON vehicles(license_plate);
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status);

-- Update sensors table to include company and vehicle context
ALTER TABLE sensors ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES companies(id) ON DELETE CASCADE;
ALTER TABLE sensors ADD COLUMN IF NOT EXISTS vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL;
ALTER TABLE sensors ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id) ON DELETE SET NULL;

-- Create esp32_devices table for better device management
CREATE TABLE IF NOT EXISTS esp32_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    device_id VARCHAR(255) UNIQUE NOT NULL, -- ESP32 MAC address or custom ID
    device_name VARCHAR(255) NOT NULL,
    firmware_version VARCHAR(50),
    hardware_revision VARCHAR(50),
    wifi_ssid VARCHAR(100),
    ip_address INET,
    mac_address VARCHAR(17),
    vehicle_id UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    installation_date DATE,
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    battery_level DECIMAL(5,2), -- Battery percentage
    signal_strength INTEGER, -- WiFi signal strength in dBm
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance', 'error')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for esp32_devices
CREATE INDEX IF NOT EXISTS idx_esp32_company_id ON esp32_devices(company_id);
CREATE INDEX IF NOT EXISTS idx_esp32_device_id ON esp32_devices(device_id);
CREATE INDEX IF NOT EXISTS idx_esp32_vehicle_id ON esp32_devices(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_esp32_status ON esp32_devices(status);
CREATE INDEX IF NOT EXISTS idx_esp32_heartbeat ON esp32_devices(last_heartbeat);

-- Create vehicle_trips table for tracking trips
CREATE TABLE IF NOT EXISTS vehicle_trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES users(id) ON DELETE SET NULL,
    helper_id UUID REFERENCES users(id) ON DELETE SET NULL,
    start_location VARCHAR(255),
    end_location VARCHAR(255),
    start_latitude DECIMAL(10,8),
    start_longitude DECIMAL(11,8),
    end_latitude DECIMAL(10,8),
    end_longitude DECIMAL(11,8),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    distance_km DECIMAL(8,2),
    duration_minutes INTEGER,
    fuel_consumption DECIMAL(8,2),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('planning', 'active', 'completed', 'cancelled')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for vehicle_trips
CREATE INDEX IF NOT EXISTS idx_trips_vehicle_id ON vehicle_trips(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_trips_driver_id ON vehicle_trips(driver_id);
CREATE INDEX IF NOT EXISTS idx_trips_start_time ON vehicle_trips(start_time);
CREATE INDEX IF NOT EXISTS idx_trips_status ON vehicle_trips(status);

-- Create company_settings table for per-company configurations
CREATE TABLE IF NOT EXISTS company_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    setting_type VARCHAR(20) DEFAULT 'string' CHECK (setting_type IN ('string', 'number', 'boolean', 'json')),
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(company_id, setting_key)
);

-- Create indexes for company_settings
CREATE INDEX IF NOT EXISTS idx_settings_company_id ON company_settings(company_id);
CREATE INDEX IF NOT EXISTS idx_settings_key ON company_settings(setting_key);

-- Create triggers for updated_at
CREATE TRIGGER update_companies_updated_at 
    BEFORE UPDATE ON companies 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at 
    BEFORE UPDATE ON teams 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicles_updated_at 
    BEFORE UPDATE ON vehicles 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_esp32_devices_updated_at 
    BEFORE UPDATE ON esp32_devices 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vehicle_trips_updated_at 
    BEFORE UPDATE ON vehicle_trips 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE companies IS 'Companies/Organizations using the IoT tracking system';
COMMENT ON TABLE teams IS 'Teams within companies (delivery teams, maintenance crews, etc.)';
COMMENT ON TABLE team_members IS 'Many-to-many relationship between teams and users';
COMMENT ON TABLE vehicles IS 'Company vehicles equipped with IoT sensors';
COMMENT ON TABLE esp32_devices IS 'ESP32 IoT devices installed in vehicles';
COMMENT ON TABLE vehicle_trips IS 'Trip records with start/end locations and metrics';
COMMENT ON TABLE company_settings IS 'Per-company configuration settings';

-- Insert master company and user (you)
INSERT INTO companies (id, name, slug, email, subscription_plan, max_users, max_vehicles, max_sensors, status) VALUES
('00000000-0000-0000-0000-000000000000', 'DashTrack Master', 'dashtrack-master', 'master@dashtrack.com', 'enterprise', 999999, 999999, 999999, 'active')
ON CONFLICT (id) DO NOTHING;

-- Create master user if not exists
INSERT INTO users (id, email, password_hash, first_name, last_name, role, company_id, status) VALUES
('11111111-1111-1111-1111-111111111111', 'master@dashtrack.com', '$2a$12$placeholder.hash.for.master.user', 'Master', 'Admin', 'master', '00000000-0000-0000-0000-000000000000', 'active')
ON CONFLICT (email) DO UPDATE SET 
    role = 'master',
    company_id = '00000000-0000-0000-0000-000000000000';

-- Insert some example companies for testing
INSERT INTO companies (name, slug, email, phone, address, city, state) VALUES
('Transportadora São Paulo', 'transportadora-sp', 'admin@transportadorasp.com', '(11) 99999-9999', 'Rua das Flores, 123', 'São Paulo', 'SP'),
('Logística Rio', 'logistica-rio', 'contato@logisticario.com', '(21) 88888-8888', 'Av. Copacabana, 456', 'Rio de Janeiro', 'RJ')
ON CONFLICT (slug) DO NOTHING;