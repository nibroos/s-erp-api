BEGIN;

CREATE TABLE IF NOT EXISTS invoice_dp_dt_boms (
  id SERIAL PRIMARY KEY,
  product_uuid TEXT,
  invoice_dp_id INT REFERENCES invoice_dps(id) ON DELETE RESTRICT,
  invoice_dp_dt_id INT REFERENCES invoice_dp_dts(id) ON DELETE RESTRICT,
  bom_id INT REFERENCES boms(id) ON DELETE RESTRICT,
  product_id INT REFERENCES products(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  product_json JSONB,
  remark TEXT,
  qty DECIMAL(20, 5),
  price DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_invoice_dp_dt_boms_invoice_dp_id ON invoice_dp_dt_boms(invoice_dp_id);

CREATE INDEX idx_invoice_dp_dt_boms_invoice_dp_dt_id ON invoice_dp_dt_boms(invoice_dp_dt_id);

CREATE INDEX idx_invoice_dp_dt_boms_bom_id ON invoice_dp_dt_boms(bom_id);

CREATE INDEX idx_invoice_dp_dt_boms_product_id ON invoice_dp_dt_boms(product_id);

CREATE INDEX idx_invoice_dp_dt_boms_item_unit_id ON invoice_dp_dt_boms(item_unit_id);

COMMIT;