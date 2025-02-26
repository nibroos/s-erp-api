BEGIN;

CREATE TABLE IF NOT EXISTS company_profiles (
  id SERIAL PRIMARY KEY,
  parent_id INT,
  is_primary INT,
  payment_id INT,
  company_name VARCHAR(255) NOT NULL,
  company_owner_name TEXT,
  company_sign_name TEXT,
  company_address TEXT,
  company_phone TEXT,
  company_email TEXT,
  company_website TEXT,
  company_sign TEXT,
  company_logo TEXT,
  company_description TEXT,
  company_remark TEXT,
  company_status INT,
  company_options_json JSONB,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_company_profiles_parent_id ON company_profiles (parent_id);

CREATE INDEX idx_company_profiles_payment_id ON company_profiles (payment_id);

COMMIT;