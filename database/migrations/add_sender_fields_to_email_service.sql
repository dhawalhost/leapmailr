-- Add sender fields to email_services table
-- Default email address for no-reply emails
DO $$ 
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'get_default_noreply_email') THEN
    CREATE FUNCTION get_default_noreply_email() RETURNS VARCHAR AS 'SELECT ''noreply@example.com''::VARCHAR' LANGUAGE SQL IMMUTABLE;
  END IF;
END $$;

ALTER TABLE email_services 
ADD COLUMN IF NOT EXISTS from_email VARCHAR(255) DEFAULT get_default_noreply_email(),
ADD COLUMN IF NOT EXISTS from_name VARCHAR(255) DEFAULT '',
ADD COLUMN IF NOT EXISTS reply_to_email VARCHAR(255) DEFAULT '';

-- Update existing records to extract from_email from configuration if available
UPDATE email_services 
SET from_email = COALESCE(
  (configuration::jsonb->>'from_email'), 
  get_default_noreply_email()
)
WHERE from_email = get_default_noreply_email() OR from_email IS NULL;

-- Make from_email NOT NULL after populating existing records
ALTER TABLE email_services ALTER COLUMN from_email SET NOT NULL;
