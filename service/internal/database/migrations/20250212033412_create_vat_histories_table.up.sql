CREATE TABLE IF NOT EXISTS vat_histories (
  id SERIAL PRIMARY KEY,
  vat_id INT NOT NULL REFERENCES mix_values (id) ON DELETE RESTRICT,
  -- num decimal
  num DECIMAL(20, 5),
  divider DECIMAL(20, 5),
  multiplier DECIMAL(20, 5),
  changed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  status INT,
  remark TEXT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_vat_histories_vat_id ON vat_histories (vat_id);