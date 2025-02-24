BEGIN;

CREATE TABLE IF NOT EXISTS customers (
  id SERIAL PRIMARY KEY,
  customer_type_id INT NOT NULL,
  agent_id INT,
  code TEXT,
  name VARCHAR(255) NOT NULL,
  address TEXT,
  phone TEXT,
  email TEXT,
  pic TEXT,
  status INT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMIT;