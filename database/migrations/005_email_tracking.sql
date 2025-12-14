-- Email Tracking Migration
-- This migration adds tables for email open and click tracking

-- Create email_trackings table
CREATE TABLE IF NOT EXISTS email_trackings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_log_id UUID NOT NULL REFERENCES email_logs(id) ON DELETE CASCADE,
    tracking_pixel_id VARCHAR(255) NOT NULL UNIQUE,
    open_count INTEGER DEFAULT 0,
    last_opened_at TIMESTAMP,
    first_opened_at TIMESTAMP,
    unique_opens INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for email_trackings
CREATE INDEX IF NOT EXISTS idx_email_trackings_email_log_id ON email_trackings(email_log_id);
CREATE INDEX IF NOT EXISTS idx_email_trackings_tracking_pixel_id ON email_trackings(tracking_pixel_id);

-- Create email_open_events table
CREATE TABLE IF NOT EXISTS email_open_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_id UUID NOT NULL REFERENCES email_trackings(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    location VARCHAR(255),
    device VARCHAR(50),
    email_client VARCHAR(100),
    opened_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for email_open_events
CREATE INDEX IF NOT EXISTS idx_email_open_events_tracking_id ON email_open_events(tracking_id);
CREATE INDEX IF NOT EXISTS idx_email_open_events_opened_at ON email_open_events(opened_at);

-- Create email_click_events table
CREATE TABLE IF NOT EXISTS email_click_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_id UUID NOT NULL REFERENCES email_trackings(id) ON DELETE CASCADE,
    link_id VARCHAR(255) NOT NULL,
    link_url TEXT NOT NULL,
    link_text VARCHAR(500),
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    location VARCHAR(255),
    device VARCHAR(50),
    clicked_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for email_click_events
CREATE INDEX IF NOT EXISTS idx_email_click_events_tracking_id ON email_click_events(tracking_id);
CREATE INDEX IF NOT EXISTS idx_email_click_events_link_id ON email_click_events(link_id);
CREATE INDEX IF NOT EXISTS idx_email_click_events_clicked_at ON email_click_events(clicked_at);

-- Add indexes to email_logs for better performance on tracking queries
CREATE INDEX IF NOT EXISTS idx_email_logs_opened_at ON email_logs(opened_at) WHERE opened_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_email_logs_clicked_at ON email_logs(clicked_at) WHERE clicked_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_email_logs_status ON email_logs(status);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_email_trackings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER email_trackings_updated_at
    BEFORE UPDATE ON email_trackings
    FOR EACH ROW
    EXECUTE FUNCTION update_email_trackings_updated_at();
