BEGIN;

CREATE TABLE IF NOT EXISTS sales_orders (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  order_type_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  warehouse_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  payment_id INT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  po_buyer_no TEXT,
  sales_order_no TEXT,
  remark TEXT,
  ship_dest TEXT,
  status TEXT DEFAULT 'WAITING',
  exchange_rate DECIMAL(20, 5),
  vat_perc DECIMAL(20, 5),
  disc_am DECIMAL(20, 5),
  disc_perc DECIMAL(20, 5),
  disc_perc_am DECIMAL(20, 5),
  disc_final DECIMAL(20, 5),
  disc_type TEXT,
  pph23_perc DECIMAL(20, 5),
  total_qty DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  total_discount DECIMAL(20, 5),
  total_pph23 DECIMAL(20, 5),
  total_vat DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  qty_out DECIMAL(20, 5),
  si_total_am DECIMAL(20, 5),
  sa_total_am DECIMAL(20, 5),
  order_at date,
  shipping_at date,
  agree_at date,
  due_at date,
  expired_at date,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN sales_orders.status IS 'WAITING, PROCESS, SHIPPED, PENDING, CANCEL, INVOICE, FINISH';

CREATE INDEX idx_sales_orders_customer_id ON sales_orders(customer_id);

CREATE INDEX idx_sales_orders_order_type_id ON sales_orders(order_type_id);

CREATE INDEX idx_sales_orders_currency_id ON sales_orders(currency_id);

CREATE INDEX idx_sales_orders_warehouse_id ON sales_orders(warehouse_id);

CREATE INDEX idx_sales_orders_vat_id ON sales_orders(vat_id);

CREATE INDEX idx_sales_orders_pph23_id ON sales_orders(pph23_id);

CREATE INDEX idx_sales_orders_branch_id ON sales_orders(branch_id);

COMMIT;