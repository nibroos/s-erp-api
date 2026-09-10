-- Add TOTP MFA support to users table.
-- Tracks the authenticator-app (RFC 6238 TOTP) credential per account alongside
-- the existing email-OTP flag (two_fa_enabled). The TOTP secret is stored
-- ENCRYPTED (AES-256-GCM ciphertext, base64) — never plaintext.
ALTER TABLE
  users
ADD
  COLUMN IF NOT EXISTS totp_enabled BOOLEAN NOT NULL DEFAULT false,
ADD
  COLUMN IF NOT EXISTS totp_secret_ciphertext TEXT,
ADD
  COLUMN IF NOT EXISTS totp_secret_key_version TEXT,
ADD
  COLUMN IF NOT EXISTS totp_status VARCHAR(16) NOT NULL DEFAULT 'NONE',
  -- NONE | PENDING | ACTIVE | REVOKED
ADD
  COLUMN IF NOT EXISTS totp_last_accepted_time_step BIGINT,
ADD
  COLUMN IF NOT EXISTS primary_mfa_method VARCHAR(16) NOT NULL DEFAULT 'none';

-- none | email | totp
-- Fast lookup of accounts with an active authenticator.
CREATE INDEX IF NOT EXISTS idx_users_totp_enabled ON users (totp_enabled)
WHERE
  totp_enabled = true;

-- Recovery (backup) codes for TOTP. Each code is single-use and stored only as
-- a SHA-256 hash; the plaintext is shown to the user exactly once at generation.
CREATE TABLE IF NOT EXISTS mfa_recovery_codes (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  code_hash TEXT NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_user ON mfa_recovery_codes (user_id);

CREATE INDEX IF NOT EXISTS idx_mfa_recovery_codes_hash ON mfa_recovery_codes (code_hash);