-- Add Multi-Factor Authentication (MFA) fields to users table (GAP-SEC-001)
-- Migration: add_mfa_fields_to_users.sql
-- Created: 2024 Phase 2 SOC 2 Compliance

-- Add MFA fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS mfa_secret TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS mfa_backup_codes TEXT DEFAULT '',
ADD COLUMN IF NOT EXISTS mfa_verified_at TIMESTAMP;

-- Create index for faster MFA lookups
CREATE INDEX IF NOT EXISTS idx_users_mfa_enabled ON users(mfa_enabled) WHERE mfa_enabled = TRUE;

-- Add comment for documentation
COMMENT ON COLUMN users.mfa_enabled IS 'Indicates if MFA is enabled for the user';
COMMENT ON COLUMN users.mfa_secret IS 'Encrypted TOTP secret for MFA (AES-256-GCM)';
COMMENT ON COLUMN users.mfa_backup_codes IS 'Encrypted array of hashed backup codes (bcrypt + AES-256-GCM)';
COMMENT ON COLUMN users.mfa_verified_at IS 'Timestamp when MFA was first enabled';

-- Audit log for migration
-- This migration adds support for TOTP-based Multi-Factor Authentication
-- Backup codes provide alternative authentication when TOTP is unavailable
-- All sensitive data (secret and backup codes) are encrypted at rest
