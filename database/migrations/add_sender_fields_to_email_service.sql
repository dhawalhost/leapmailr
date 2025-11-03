-- Add sender fields to email_services table
ALTER TABLE email_services 
ADD COLUMN IF NOT EXISTS from_email VARCHAR(255) DEFAULT 'noreply@example.com',
ADD COLUMN IF NOT EXISTS from_name VARCHAR(255) DEFAULT '',
ADD COLUMN IF NOT EXISTS reply_to_email VARCHAR(255) DEFAULT '';

-- Update existing records to extract from_email from configuration if available
UPDATE email_services 
SET from_email = COALESCE(
  (configuration::jsonb->>'from_email'), 
  'noreply@example.com'
)
WHERE from_email = 'noreply@example.com' OR from_email IS NULL;

-- Make from_email NOT NULL after populating existing records
ALTER TABLE email_services ALTER COLUMN from_email SET NOT NULL;
