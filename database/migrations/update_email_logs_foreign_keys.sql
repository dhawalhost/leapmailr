-- Migration: Update email_logs foreign key constraints to allow template/service deletion
-- Date: 2025-11-01
-- Description: Change foreign key constraints from RESTRICT to SET NULL for template_id and service_id
--              This allows templates and services to be deleted without being blocked by email logs

-- Drop existing foreign key constraints (GORM-generated names)
ALTER TABLE email_logs DROP CONSTRAINT IF EXISTS fk_templates_email_logs;
ALTER TABLE email_logs DROP CONSTRAINT IF EXISTS fk_email_services_email_logs;

-- Recreate foreign key constraints with ON DELETE SET NULL
ALTER TABLE email_logs 
    ADD CONSTRAINT fk_templates_email_logs 
    FOREIGN KEY (template_id) 
    REFERENCES templates(id) 
    ON DELETE SET NULL;

ALTER TABLE email_logs 
    ADD CONSTRAINT fk_email_services_email_logs 
    FOREIGN KEY (service_id) 
    REFERENCES email_services(id) 
    ON DELETE SET NULL;

-- Add comment for documentation
COMMENT ON CONSTRAINT fk_templates_email_logs ON email_logs IS 'Foreign key to templates with SET NULL on delete - preserves email log history when template is deleted';
COMMENT ON CONSTRAINT fk_email_services_email_logs ON email_logs IS 'Foreign key to email_services with SET NULL on delete - preserves email log history when service is deleted';
