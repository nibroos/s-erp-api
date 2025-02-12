BEGIN;

CREATE TABLE IF NOT EXISTS vat_histories (
  id SERIAL PRIMARY KEY,
  vat_id INT NOT NULL,
  num INT,
  divider INT,
  multiplier INT,
  changed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  status INT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMIT;