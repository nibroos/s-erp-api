CREATE TABLE IF NOT EXISTS request_order_dts (
  id SERIAL PRIMARY KEY,
  product_uuid TEXT,
  request_order_id INT REFERENCES request_orders(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  ref_id INT,
  product_id INT,
  item_id INT,
  ref_type TEXT,
  ref_json JSONB,
  product_type TEXT,
  product_name TEXT,
  item_name TEXT,
  unit_name TEXT,
  price_sell DECIMAL(20, 5),
  remark TEXT,
  product_json JSONB,
  order_product_qty DECIMAL(20, 5),
  order_item_qty DECIMAL(20, 5),
  wh_qty DECIMAL(20, 5),
  req_qty DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_request_order_dts_request_order_id ON request_order_dts(request_order_id);

CREATE INDEX idx_request_order_dts_item_unit_id ON request_order_dts(item_unit_id);

CREATE INDEX idx_request_order_dts_ref_id ON request_order_dts(ref_id);

CREATE INDEX idx_request_order_dts_product_id ON request_order_dts(product_id);

CREATE INDEX idx_request_order_dts_item_id ON request_order_dts(item_id);