BEGIN;

CREATE TABLE IF NOT EXISTS invoice_dp_dts (
  id SERIAL PRIMARY KEY,
  product_uuid TEXT,
  invoice_dp_id INT REFERENCES invoice_dps(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  ref_id INT,
  ref_dt_id INT,
  product_id INT,
  ref_type TEXT,
  ref_json JSONB,
  product_type TEXT,
  product_json JSONB,
  remark TEXT,
  dp_percentage DECIMAL(20, 5),
  is_vat INT DEFAULT 0,
  is_pph23 INT DEFAULT 0,
  qty DECIMAL(20, 5),
  price DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  discount_amount DECIMAL(20, 5),
  discount_percentage DECIMAL(20, 5),
  discount_percentage_num DECIMAL(20, 5),
  discount_percentage_amount DECIMAL(20, 5),
  discount_final DECIMAL(20, 5),
  discount_type TEXT,
  total_amount DECIMAL(20, 5),
  total_dp DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_invoice_dp_dts_invoice_dp_id ON invoice_dp_dts(invoice_dp_id);

CREATE INDEX idx_invoice_dp_dts_item_unit_id ON invoice_dp_dts(item_unit_id);

CREATE INDEX idx_invoice_dp_dts_ref_id ON invoice_dp_dts(ref_id);

CREATE INDEX idx_invoice_dp_dts_vat_id ON invoice_dp_dts(vat_id);

CREATE INDEX idx_invoice_dp_dts_pph23_id ON invoice_dp_dts(pph23_id);

CREATE INDEX idx_invoice_dp_dts_product_id ON invoice_dp_dts(product_id);

COMMIT;