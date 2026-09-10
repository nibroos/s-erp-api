-- Remove two_fa_enabled column from users table
DROP INDEX IF EXISTS idx_users_two_fa_enabled;

ALTER TABLE
  users DROP COLUMN IF EXISTS two_fa_enabled;