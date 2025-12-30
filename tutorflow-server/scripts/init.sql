-- Init script for PostgreSQL
-- This runs automatically when the container first starts

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create custom types
DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('admin', 'manager', 'tutor', 'student');

EXCEPTION WHEN duplicate_object THEN null;

END $$;

DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('active', 'inactive', 'suspended', 'pending');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Add more enum types as needed...
-- The application will handle creating the rest via GORM AutoMigrate