BEGIN;

CREATE TABLE IF NOT EXISTS quo_dt_boms (
  id SERIAL PRIMARY KEY,
  quotation_id INT REFERENCES quotations(id) ON DELETE RESTRICT,
  quo_dt_id INT REFERENCES quo_dts(id) ON DELETE RESTRICT,
  product_id INT REFERENCES products(id) ON DELETE RESTRICT,
  item_id INT REFERENCES products(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  ref_json JSONB,
  gen_code TEXT,
  remark TEXT,
  qty DECIMAL(20, 5),
  price_buy DECIMAL(20, 5),
  price_sell DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_quo_dt_boms_quotation_id ON quo_dt_boms(quotation_id);

CREATE INDEX idx_quo_dt_boms_quo_dt_id ON quo_dt_boms(quo_dt_id);

CREATE INDEX idx_quo_dt_boms_product_id ON quo_dt_boms(product_id);

CREATE INDEX idx_quo_dt_boms_item_id ON quo_dt_boms(item_id);

CREATE INDEX idx_quo_dt_boms_item_unit_id ON quo_dt_boms(item_unit_id);

COMMIT;