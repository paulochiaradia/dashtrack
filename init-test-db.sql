-- Initialize test database
-- This script is executed when the PostgreSQL container starts

-- Create test databases
CREATE DATABASE dashtrack_test_unit;
CREATE DATABASE dashtrack_test_integration;
CREATE DATABASE dashtrack_test_hierarchy;

-- Grant permissions
GRANT ALL PRIVILEGES ON DATABASE dashtrack_test_unit TO dashtrack_user;
GRANT ALL PRIVILEGES ON DATABASE dashtrack_test_integration TO dashtrack_user;
GRANT ALL PRIVILEGES ON DATABASE dashtrack_test_hierarchy TO dashtrack_user;

-- Create extensions for each test database
\c dashtrack_test_unit;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c dashtrack_test_integration;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c dashtrack_test_hierarchy;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";