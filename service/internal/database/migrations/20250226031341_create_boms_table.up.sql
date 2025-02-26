BEGIN;

CREATE TABLE IF NOT EXISTS boms (
  id SERIAL PRIMARY KEY,
  product_id INT NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
  ms_item_id INT NOT NULL REFERENCES ms_items (id) ON DELETE RESTRICT,
  item_unit_id INT NOT NULL REFERENCES item_units (id) ON DELETE RESTRICT,
  qty DECIMAL(20, 5),
  remark TEXT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_boms_ms_item_id ON boms(ms_item_id);

CREATE INDEX idx_boms_item_unit_id ON boms(item_unit_id);

COMMIT;