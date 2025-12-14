-- Migration: Update foreign key constraint on email_logs to SET NULL on delete
-- Reason: Allow deleting email services without deleting email logs (preserve history)
-- Date: 2025-11-05

-- Drop the existing foreign key constraint
ALTER TABLE email_logs 
DROP CONSTRAINT IF EXISTS fk_email_services_email_logs;

-- Add the new constraint with ON DELETE SET NULL
ALTER TABLE email_logs
ADD CONSTRAINT fk_email_services_email_logs 
FOREIGN KEY (service_id) 
REFERENCES email_services(id) 
ON DELETE SET NULL;

-- Add comment for clarity
COMMENT ON CONSTRAINT fk_email_services_email_logs ON email_logs IS 'Sets service_id to NULL when email service is deleted to preserve email history';
