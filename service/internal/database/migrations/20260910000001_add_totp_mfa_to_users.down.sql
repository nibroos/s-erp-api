-- Roll back TOTP MFA support.
DROP TABLE IF EXISTS mfa_recovery_codes;

ALTER TABLE
  users DROP COLUMN IF EXISTS totp_enabled,
  DROP COLUMN IF EXISTS totp_secret_ciphertext,
  DROP COLUMN IF EXISTS totp_secret_key_version,
  DROP COLUMN IF EXISTS totp_status,
  DROP COLUMN IF EXISTS totp_last_accepted_time_step,
  DROP COLUMN IF EXISTS primary_mfa_method;