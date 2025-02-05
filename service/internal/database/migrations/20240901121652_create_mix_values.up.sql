BEGIN;

CREATE TABLE IF NOT EXISTS mix_values (
  id SERIAL PRIMARY KEY,
  group_id INT NOT NULL,
  parent_id INT NOT NULL,
  name VARCHAR(500) NOT NULL,
  description TEXT,
  remark TEXT,
  VALUE DECIMAL(20, 5),
  status INT,
  options_json JSONB,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMIT;