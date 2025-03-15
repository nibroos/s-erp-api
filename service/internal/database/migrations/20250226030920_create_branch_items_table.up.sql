BEGIN;

CREATE TABLE IF NOT EXISTS branch_items (
  id SERIAL PRIMARY KEY,
  branch_id INT NOT NULL REFERENCES branches (id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units (id) ON DELETE RESTRICT,
  code TEXT,
  factory_code TEXT,
  name VARCHAR(255) NOT NULL,
  sku TEXT,
  barcode TEXT,
  specification TEXT,
  description TEXT,
  remark TEXT,
  tpb_code TEXT,
  minimum_stock DECIMAL(20, 5),
  qty_stock DECIMAL(20, 5) DEFAULT 0,
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

CREATE INDEX idx_branch_items_branch_id ON branch_items(branch_id);

CREATE INDEX idx_branch_items_item_unit_id ON branch_items(item_unit_id);

COMMIT;