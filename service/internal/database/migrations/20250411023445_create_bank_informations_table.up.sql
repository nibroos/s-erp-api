BEGIN;

CREATE TABLE IF NOT EXISTS bank_informations (
  id SERIAL PRIMARY KEY,
  commpany_profile_id INT REFERENCES company_profiles(id) ON DELETE RESTRICT,
  name TEXT,
  account_number TEXT,
  account_name TEXT,
  description TEXT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_bank_informations_commpany_profile_id ON bank_informations(commpany_profile_id);

COMMIT;