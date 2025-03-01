BEGIN;

CREATE TABLE IF NOT EXISTS quotations (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  order_type_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  payment_id INT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  quo_no TEXT,
  title TEXT,
  remark TEXT,
  status TEXT,
  exchange_rate DECIMAL(20, 5),
  vat_perc DECIMAL(20, 5),
  pph23_perc DECIMAL(20, 5),
  total_qty DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  total_discount DECIMAL(20, 5),
  total_pph23 DECIMAL(20, 5),
  total_vat DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  due_at timestamp with time zone,
  expired_at timestamp with time zone,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_quotations_customer_id ON quotations(customer_id);

CREATE INDEX idx_quotations_order_type_id ON quotations(order_type_id);

CREATE INDEX idx_quotations_currency_id ON quotations(currency_id);

CREATE INDEX idx_quotations_vat_id ON quotations(vat_id);

CREATE INDEX idx_quotations_pph23_id ON quotations(pph23_id);

CREATE INDEX idx_quotations_branch_id ON quotations(branch_id);

COMMIT;