-- +migrate Up
-- Create vehicles table for fleet management
CREATE TABLE vehicles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    
    -- Vehicle identification
    license_plate VARCHAR(20) NOT NULL,
    brand VARCHAR(100) NOT NULL,
    model VARCHAR(100) NOT NULL,
    year INTEGER NOT NULL,
    color VARCHAR(50),
    vin VARCHAR(17), -- Vehicle Identification Number
    
    -- Vehicle classification
    vehicle_type VARCHAR(50) NOT NULL CHECK (vehicle_type IN ('car', 'van', 'truck', 'motorcycle', 'bus', 'other')),
    fuel_type VARCHAR(50) NOT NULL CHECK (fuel_type IN ('gasoline', 'diesel', 'electric', 'hybrid', 'ethanol', 'cng')),
    
    -- Technical details
    engine_capacity VARCHAR(20),
    transmission VARCHAR(20) CHECK (transmission IN ('manual', 'automatic', 'semi-automatic')),
    seats INTEGER,
    cargo_capacity DECIMAL(10, 2), -- in cubic meters or kg
    
    -- Ownership & Registration
    ownership VARCHAR(20) CHECK (ownership IN ('owned', 'leased', 'rented')),
    registration_number VARCHAR(50),
    registration_expiry DATE,
    
    -- Insurance
    insurance_company VARCHAR(100),
    insurance_policy VARCHAR(50),
    insurance_expiry DATE,
    
    -- Maintenance
    last_maintenance_date DATE,
    next_maintenance_date DATE,
    odometer INTEGER, -- current mileage in km
    
    -- Operational status
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'maintenance', 'retired')),
    availability VARCHAR(50) DEFAULT 'available' CHECK (availability IN ('available', 'in_use', 'maintenance', 'unavailable')),
    
    -- GPS/Tracking
    gps_device_id VARCHAR(100),
    last_location_lat DECIMAL(10, 8),
    last_location_lng DECIMAL(11, 8),
    last_location_updated TIMESTAMP WITH TIME ZONE,
    
    -- Additional info
    notes TEXT,
    images JSONB DEFAULT '[]'::jsonb, -- Array of image URLs
    documents JSONB DEFAULT '[]'::jsonb, -- Array of document references
    
    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT vehicles_company_plate_unique UNIQUE (company_id, license_plate, deleted_at)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_vehicles_company_id ON vehicles(company_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_license_plate ON vehicles(license_plate) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_status ON vehicles(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_availability ON vehicles(availability) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_vehicle_type ON vehicles(vehicle_type) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_created_at ON vehicles(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_vehicles_gps_device ON vehicles(gps_device_id) WHERE gps_device_id IS NOT NULL;

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_vehicles_company_status ON vehicles(company_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vehicles_company_availability ON vehicles(company_id, availability) WHERE deleted_at IS NULL;

-- Add comments for documentation
COMMENT ON TABLE vehicles IS 'Fleet vehicles managed by companies';
COMMENT ON COLUMN vehicles.license_plate IS 'Vehicle license plate number (unique per company)';
COMMENT ON COLUMN vehicles.vin IS 'Vehicle Identification Number (17 characters)';
COMMENT ON COLUMN vehicles.vehicle_type IS 'Type of vehicle: car, van, truck, motorcycle, bus, other';
COMMENT ON COLUMN vehicles.fuel_type IS 'Fuel type: gasoline, diesel, electric, hybrid, ethanol, cng';
COMMENT ON COLUMN vehicles.status IS 'Operational status: active, inactive, maintenance, retired';
COMMENT ON COLUMN vehicles.availability IS 'Current availability: available, in_use, maintenance, unavailable';
COMMENT ON COLUMN vehicles.ownership IS 'Ownership type: owned, leased, rented';
COMMENT ON COLUMN vehicles.images IS 'JSON array of vehicle image URLs';
COMMENT ON COLUMN vehicles.documents IS 'JSON array of vehicle document references';
