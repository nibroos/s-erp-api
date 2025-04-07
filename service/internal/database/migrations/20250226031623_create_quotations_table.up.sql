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
  rev_no INT DEFAULT 0,
  quo_no TEXT,
  title TEXT,
  remark TEXT,
  status TEXT DEFAULT 'WAITING',
  exchange_rate DECIMAL(20, 5),
  vat_perc DECIMAL(20, 5),
  pph23_perc DECIMAL(20, 5),
  markup_perc DECIMAL(20, 5),
  is_vat INT,
  is_pph23 INT,
  disc_am DECIMAL(20, 5),
  disc_perc DECIMAL(20, 5),
  disc_perc_am DECIMAL(20, 5),
  disc_final DECIMAL(20, 5),
  disc_type TEXT,
  total_qty DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  total_discount DECIMAL(20, 5),
  total_pph23 DECIMAL(20, 5),
  total_vat DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  due_at date,
  expired_at date,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN quotations.status IS 'WAITING, APPROVED, PENDING, CANCELED';

CREATE INDEX idx_quotations_customer_id ON quotations(customer_id);

CREATE INDEX idx_quotations_order_type_id ON quotations(order_type_id);

CREATE INDEX idx_quotations_currency_id ON quotations(currency_id);

CREATE INDEX idx_quotations_vat_id ON quotations(vat_id);

CREATE INDEX idx_quotations_pph23_id ON quotations(pph23_id);

CREATE INDEX idx_quotations_branch_id ON quotations(branch_id);

COMMIT;