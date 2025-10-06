-- Migration: Create IoT Sensor Tables
-- Date: 2025-09-29
-- Description: Create tables for ESP32 sensor data management

-- Create sensors table
CREATE TABLE IF NOT EXISTS sensors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('dht11', 'gyroscope', 'gps_neo6v2', 'generic')),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'error')),
    location VARCHAR(255),
    description TEXT,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP WITH TIME ZONE
);

-- Create indexes for sensors
CREATE INDEX IF NOT EXISTS idx_sensors_device_id ON sensors(device_id);
CREATE INDEX IF NOT EXISTS idx_sensors_user_id ON sensors(user_id);
CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(type);
CREATE INDEX IF NOT EXISTS idx_sensors_status ON sensors(status);
CREATE INDEX IF NOT EXISTS idx_sensors_last_seen ON sensors(last_seen);

-- Create DHT11 readings table
CREATE TABLE IF NOT EXISTS dht11_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    temperature DECIMAL(5,2) NOT NULL, -- -99.99 to 99.99 Celsius
    humidity DECIMAL(5,2) NOT NULL CHECK (humidity >= 0 AND humidity <= 100), -- 0-100%
    heat_index DECIMAL(5,2), -- Calculated heat index
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for DHT11 readings
CREATE INDEX IF NOT EXISTS idx_dht11_sensor_id ON dht11_readings(sensor_id);
CREATE INDEX IF NOT EXISTS idx_dht11_device_id ON dht11_readings(device_id);
CREATE INDEX IF NOT EXISTS idx_dht11_timestamp ON dht11_readings(timestamp);
CREATE INDEX IF NOT EXISTS idx_dht11_created_at ON dht11_readings(created_at);

-- Create gyroscope readings table
CREATE TABLE IF NOT EXISTS gyroscope_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    accel_x DECIMAL(10,6) NOT NULL, -- m/s²
    accel_y DECIMAL(10,6) NOT NULL, -- m/s²
    accel_z DECIMAL(10,6) NOT NULL, -- m/s²
    gyro_x DECIMAL(10,6) NOT NULL,  -- rad/s
    gyro_y DECIMAL(10,6) NOT NULL,  -- rad/s
    gyro_z DECIMAL(10,6) NOT NULL,  -- rad/s
    magnitude DECIMAL(10,6) NOT NULL, -- Overall acceleration magnitude
    is_vibrating BOOLEAN NOT NULL DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for gyroscope readings
CREATE INDEX IF NOT EXISTS idx_gyroscope_sensor_id ON gyroscope_readings(sensor_id);
CREATE INDEX IF NOT EXISTS idx_gyroscope_device_id ON gyroscope_readings(device_id);
CREATE INDEX IF NOT EXISTS idx_gyroscope_timestamp ON gyroscope_readings(timestamp);
CREATE INDEX IF NOT EXISTS idx_gyroscope_vibrating ON gyroscope_readings(is_vibrating);
CREATE INDEX IF NOT EXISTS idx_gyroscope_created_at ON gyroscope_readings(created_at);

-- Create GPS readings table
CREATE TABLE IF NOT EXISTS gps_readings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,  -- -90.00000000 to 90.00000000
    longitude DECIMAL(11,8) NOT NULL, -- -180.00000000 to 180.00000000
    altitude DECIMAL(8,2),            -- Meters above sea level
    speed DECIMAL(8,2) DEFAULT 0,     -- km/h
    heading DECIMAL(6,2),             -- 0-359.99 degrees from North
    satellites INTEGER DEFAULT 0,     -- Number of satellites
    hdop DECIMAL(5,2),               -- Horizontal Dilution of Precision
    is_valid BOOLEAN NOT NULL DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for GPS readings
CREATE INDEX IF NOT EXISTS idx_gps_sensor_id ON gps_readings(sensor_id);
CREATE INDEX IF NOT EXISTS idx_gps_device_id ON gps_readings(device_id);
CREATE INDEX IF NOT EXISTS idx_gps_timestamp ON gps_readings(timestamp);
CREATE INDEX IF NOT EXISTS idx_gps_location ON gps_readings(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_gps_valid ON gps_readings(is_valid);
CREATE INDEX IF NOT EXISTS idx_gps_created_at ON gps_readings(created_at);

-- Create sensor alerts table
CREATE TABLE IF NOT EXISTS sensor_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sensor_id UUID NOT NULL REFERENCES sensors(id) ON DELETE CASCADE,
    type VARCHAR(100) NOT NULL, -- temperature_high, vibration_detected, gps_out_of_bounds, etc.
    message TEXT NOT NULL,
    value DECIMAL(15,6),
    threshold DECIMAL(15,6),
    severity VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for sensor alerts
CREATE INDEX IF NOT EXISTS idx_alerts_sensor_id ON sensor_alerts(sensor_id);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON sensor_alerts(type);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON sensor_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON sensor_alerts(created_at);
CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON sensor_alerts(resolved_at);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for sensors table
CREATE TRIGGER update_sensors_updated_at 
    BEFORE UPDATE ON sensors 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE sensors IS 'ESP32 sensor devices registered in the system';
COMMENT ON TABLE dht11_readings IS 'Temperature and humidity readings from DHT11 sensors';
COMMENT ON TABLE gyroscope_readings IS 'Acceleration and gyroscope readings for vibration detection';
COMMENT ON TABLE gps_readings IS 'GPS location readings from NEO-6V2 modules';
COMMENT ON TABLE sensor_alerts IS 'Automated alerts based on sensor thresholds';

-- Insert example sensor configurations (optional)
-- INSERT INTO sensors (device_id, name, type, location, description, user_id) VALUES
-- ('ESP32_DHT11_001', 'Sala Principal - Temperatura', 'dht11', 'Sala de estar', 'Sensor de temperatura e umidade da sala principal', 
--  (SELECT id FROM users WHERE email = 'admin@dashtrack.com' LIMIT 1));