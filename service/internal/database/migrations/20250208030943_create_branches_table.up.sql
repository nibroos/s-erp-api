CREATE TABLE IF NOT EXISTS branches (
  id SERIAL PRIMARY KEY,
  parent_id INT,
  company_profile_id INT REFERENCES company_profiles(id) ON DELETE RESTRICT,
  name VARCHAR(255) NOT NULL,
  owner_name TEXT,
  sign_name TEXT,
  address TEXT,
  phone TEXT,
  email TEXT,
  website TEXT,
  sign TEXT,
  logo TEXT,
  description TEXT,
  remark TEXT,
  status INT,
  options_json JSONB,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_branches_parent_id ON branches (parent_id);

CREATE INDEX idx_branches_company_profile_id ON branches (company_profile_id);