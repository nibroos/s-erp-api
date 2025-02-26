BEGIN;

CREATE TABLE IF NOT EXISTS catalogs (
  id SERIAL PRIMARY KEY,
  unit_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  collection_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  code TEXT,
  factory_code TEXT,
  name VARCHAR(255) NOT NULL,
  sku TEXT,
  barcode TEXT,
  specification TEXT,
  description TEXT,
  remark TEXT,
  price_sell DECIMAL(20, 5),
  price_buy DECIMAL(20, 5),
  margin DECIMAL(20, 5),
  status INT,
  expired_at timestamp with time zone,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_catalogs_unit_id ON catalogs(unit_id);

CREATE INDEX idx_catalogs_collection_id ON catalogs(collection_id);

CREATE INDEX idx_catalogs_branch_id ON catalogs(branch_id);

COMMIT;