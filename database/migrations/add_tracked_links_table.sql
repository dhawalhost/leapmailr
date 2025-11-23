-- Add TrackedLink table for secure link tracking
-- SECURITY: Prevents open redirect vulnerabilities by storing pre-approved URLs

-- Create tracked_links table
CREATE TABLE IF NOT EXISTS tracked_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tracking_pixel_id VARCHAR(255) NOT NULL,
    link_id VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Composite index for fast lookups during redirect
    CONSTRAINT idx_tracking_link UNIQUE (tracking_pixel_id, link_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_tracked_links_tracking_pixel_id ON tracked_links(tracking_pixel_id);
CREATE INDEX IF NOT EXISTS idx_tracked_links_link_id ON tracked_links(link_id);

-- Add comments for documentation
COMMENT ON TABLE tracked_links IS 'Stores pre-approved URLs for email link tracking to prevent open redirect vulnerabilities';
COMMENT ON COLUMN tracked_links.tracking_pixel_id IS 'References the email tracking pixel';
COMMENT ON COLUMN tracked_links.link_id IS 'Unique identifier for the link within the email';
COMMENT ON COLUMN tracked_links.original_url IS 'The pre-approved destination URL (stored at email send time)';
