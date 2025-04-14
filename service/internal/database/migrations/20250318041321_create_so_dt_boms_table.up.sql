CREATE TABLE IF NOT EXISTS so_dt_boms (
  id SERIAL PRIMARY KEY,
  product_uuid TEXT,
  sales_order_id INT REFERENCES sales_orders(id) ON DELETE RESTRICT,
  so_dt_id INT REFERENCES so_dts(id) ON DELETE RESTRICT,
  product_id INT REFERENCES products(id) ON DELETE RESTRICT,
  item_id INT REFERENCES products(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  item_json JSONB,
  gen_code TEXT,
  remark TEXT,
  qty DECIMAL(20, 5),
  qty_out DECIMAL(20, 5),
  price_sell DECIMAL(20, 5),
  price_buy DECIMAL(20, 5),
  subtotal_sell DECIMAL(20, 5),
  subtotal_buy DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_so_dt_boms_sales_order_id ON so_dt_boms(sales_order_id);

CREATE INDEX idx_so_dt_boms_so_dt_id ON so_dt_boms(so_dt_id);

CREATE INDEX idx_so_dt_boms_product_id ON so_dt_boms(product_id);

CREATE INDEX idx_so_dt_boms_item_id ON so_dt_boms(item_id);

CREATE INDEX idx_so_dt_boms_item_unit_id ON so_dt_boms(item_unit_id);