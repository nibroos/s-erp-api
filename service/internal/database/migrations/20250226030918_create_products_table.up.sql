BEGIN;

CREATE TABLE IF NOT EXISTS products (
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

CREATE INDEX idx_products_unit_id ON products(unit_id);

CREATE INDEX idx_products_collection_id ON products(collection_id);

CREATE INDEX idx_products_branch_id ON products(branch_id);

COMMIT;