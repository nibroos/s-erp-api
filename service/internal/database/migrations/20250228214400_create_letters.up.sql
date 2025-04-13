CREATE TABLE IF NOT EXISTS letters (
  id SERIAL PRIMARY KEY,
  ref_id INT,
  ref_type TEXT,
  file_type TEXT,
  file_url TEXT,
  file_name TEXT,
  remark TEXT,
  file_prop JSONB,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);