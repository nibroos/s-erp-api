BEGIN;

CREATE TABLE IF NOT EXISTS company_profiles (
  id SERIAL PRIMARY KEY,
  parent_id INT,
  is_primary INT,
  company_name VARCHAR(255) NOT NULL,
  company_address TEXT,
  company_phone TEXT,
  company_email TEXT,
  company_website TEXT,
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

COMMIT;