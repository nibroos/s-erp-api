-- Add two_fa_enabled column to users table for per-account 2FA control
-- This allows users to enable/disable 2FA individually instead of relying on global env config
ALTER TABLE
  users
ADD
  COLUMN IF NOT EXISTS two_fa_enabled BOOLEAN NOT NULL DEFAULT false;

-- Add index for faster lookups when checking 2FA status during login
CREATE INDEX IF NOT EXISTS idx_users_two_fa_enabled ON users (two_fa_enabled)
WHERE
  two_fa_enabled = true;