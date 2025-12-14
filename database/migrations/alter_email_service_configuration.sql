-- Migration: Change email_services.configuration column from JSONB to TEXT
-- Reason: The configuration is stored as encrypted text, not JSON
-- Date: 2025-11-05

-- Alter the column type from JSONB to TEXT
ALTER TABLE email_services 
ALTER COLUMN configuration TYPE TEXT USING configuration::TEXT;

-- Add comment for clarity
COMMENT ON COLUMN email_services.configuration IS 'Encrypted configuration data stored as base64 text';
