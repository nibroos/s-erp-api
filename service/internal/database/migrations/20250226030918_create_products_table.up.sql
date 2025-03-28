BEGIN;

CREATE TABLE IF NOT EXISTS products (
  id SERIAL PRIMARY KEY,
  item_sub_group_id INT NOT NULL REFERENCES mix_values(id) ON DELETE RESTRICT,
  item_unit_id INT,
  code TEXT,
  factory_code TEXT,
  name VARCHAR(255) NOT NULL,
  prod_type TEXT DEFAULT 'single',
  sku TEXT,
  barcode TEXT,
  specification TEXT,
  description TEXT,
  remark TEXT,
  tpb_code TEXT,
  minimum_stock DECIMAL(20, 5),
  qty_stock DECIMAL(20, 5) DEFAULT 0,
  is_all_branch INT,
  is_vat INT DEFAULT 0,
  is_pph23 INT DEFAULT 0,
  status INT,
  expired_at timestamp with time zone,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN products.prod_type IS 'product, single';

CREATE INDEX idx_products_item_sub_group_id ON products(item_sub_group_id);

CREATE INDEX idx_products_item_unit_id ON products(item_unit_id);

COMMIT;