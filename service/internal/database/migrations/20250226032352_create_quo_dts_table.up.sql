BEGIN;

CREATE TABLE IF NOT EXISTS quo_dts (
  id SERIAL PRIMARY KEY,
  quotation_id INT REFERENCES quotations(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  ref_id INT,
  item_id INT,
  ref_json JSONB,
  ref_type TEXT,
  item_type TEXT,
  item_json JSONB,
  remark TEXT,
  vat_perc DECIMAL(20, 5),
  qty_so DECIMAL(20, 5),
  qty DECIMAL(20, 5),
  price_sell DECIMAL(20, 5),
  price_buy DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  disc_am DECIMAL(20, 5),
  disc_perc DECIMAL(20, 5),
  total_am DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN quo_dts.ref_type IS 'master';

COMMENT ON COLUMN quo_dts.item_type IS 'item, product';

CREATE INDEX idx_quo_dts_quotation_id ON quo_dts(quotation_id);

CREATE INDEX idx_quo_dts_item_unit_id ON quo_dts(item_unit_id);

CREATE INDEX idx_quo_dts_vat_id ON quo_dts(vat_id);

CREATE INDEX idx_quo_dts_ref_id ON quo_dts(ref_id);

COMMIT;