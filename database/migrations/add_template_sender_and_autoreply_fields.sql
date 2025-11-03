-- Migration: Add sender override and auto-reply fields to templates table
-- Date: 2025-11-01
-- Description: Allows templates to override service sender email/name and configure auto-reply

-- Add sender override fields
ALTER TABLE templates ADD COLUMN IF NOT EXISTS from_email VARCHAR(255);
ALTER TABLE templates ADD COLUMN IF NOT EXISTS from_name VARCHAR(255);
ALTER TABLE templates ADD COLUMN IF NOT EXISTS reply_to_email VARCHAR(255);

-- Add auto-reply fields
ALTER TABLE templates ADD COLUMN IF NOT EXISTS auto_reply_enabled BOOLEAN DEFAULT false;
ALTER TABLE templates ADD COLUMN IF NOT EXISTS auto_reply_template_id UUID;

-- Add foreign key constraint for auto_reply_template_id
ALTER TABLE templates ADD CONSTRAINT fk_auto_reply_template
    FOREIGN KEY (auto_reply_template_id) 
    REFERENCES templates(id) 
    ON DELETE SET NULL;

-- Add comment for documentation
COMMENT ON COLUMN templates.from_email IS 'Override service from_email when set';
COMMENT ON COLUMN templates.from_name IS 'Override service from_name when set';
COMMENT ON COLUMN templates.reply_to_email IS 'Override service reply_to_email when set';
COMMENT ON COLUMN templates.auto_reply_enabled IS 'Enable automatic reply when email is sent using this template';
COMMENT ON COLUMN templates.auto_reply_template_id IS 'Template to use for auto-reply (must be an auto_reply category template)';

-- Update category enum to include auto_reply (if using enum type)
-- Note: PostgreSQL doesn't allow adding values to enum in a transaction-safe way
-- If you're using CHECK constraints instead of ENUM, update accordingly
