CREATE TABLE IF NOT EXISTS so_dts (
  id SERIAL PRIMARY KEY,
  product_uuid TEXT,
  sales_order_id INT REFERENCES sales_orders(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  ref_id INT,
  item_id INT,
  ref_json JSONB,
  ref_type TEXT,
  item_type TEXT,
  item_json JSONB,
  gen_code TEXT,
  remark TEXT,
  -- is_lock_vat INT,
  vat_perc DECIMAL(20, 5),
  vat_perc_am DECIMAL(20, 5),
  -- is_lock_pph23 INT,
  pph23_perc DECIMAL(20, 5),
  pph23_perc_am DECIMAL(20, 5),
  markup_perc DECIMAL(20, 5),
  markup_perc_am DECIMAL(20, 5),
  is_vat INT DEFAULT 0,
  is_pph23 INT DEFAULT 0,
  is_lock_markup INT,
  is_lock_price_sell INT,
  qty_out DECIMAL(20, 5),
  qty DECIMAL(20, 5),
  price_sell DECIMAL(20, 5),
  price_buy DECIMAL(20, 5),
  subtotal_sell DECIMAL(20, 5),
  subtotal_buy DECIMAL(20, 5),
  disc_am DECIMAL(20, 5),
  disc_perc DECIMAL(20, 5),
  disc_perc_num DECIMAL(20, 5),
  disc_perc_am DECIMAL(20, 5),
  disc_final DECIMAL(20, 5),
  disc_type TEXT,
  -- head_disc_am DECIMAL(20, 5),
  -- head_disc_perc_am DECIMAL(20, 5),
  -- disc_end DECIMAL(20, 5),
  total_am DECIMAL(20, 5),
  si_total_am DECIMAL(20, 5),
  sa_total_am DECIMAL(20, 5),
  total_dp DECIMAL(20, 5),
  history_total_dp DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN quo_dts.ref_type IS 'products, quotations';

COMMENT ON COLUMN quo_dts.item_type IS 'item, product';

COMMENT ON COLUMN quo_dts.disc_type IS 'p = %, a = amount';

CREATE INDEX idx_so_dts_sales_order_id ON so_dts(sales_order_id);

CREATE INDEX idx_so_dts_item_unit_id ON so_dts(item_unit_id);

CREATE INDEX idx_so_dts_ref_id ON so_dts(ref_id);

CREATE INDEX idx_so_dts_vat_id ON so_dts(vat_id);

CREATE INDEX idx_so_dts_pph23_id ON so_dts(pph23_id);