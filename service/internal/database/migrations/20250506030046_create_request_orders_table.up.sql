CREATE TABLE IF NOT EXISTS request_orders (
  id SERIAL PRIMARY KEY,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  warehouse_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  request_no TEXT,
  request_date date,
  remark TEXT,
  requested TEXT,
  rev_no INT,
  status TEXT,
  grand_total_order_product_qty DECIMAL(20, 5),
  grand_total_order_item_qty DECIMAL(20, 5),
  grand_total_wh_qty DECIMAL(20, 5),
  grand_total_req_qty DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_request_orders_branch_id ON request_orders(branch_id);

CREATE INDEX idx_request_orders_warehouse_id ON request_orders(warehouse_id);