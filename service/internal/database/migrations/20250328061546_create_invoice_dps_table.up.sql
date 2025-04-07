BEGIN;

CREATE TABLE IF NOT EXISTS invoice_dps (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  payment_term_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  invoice_no TEXT,
  invoice_date date,
  exchange_rate DECIMAL(20, 5),
  remark TEXT,
  status TEXT,
  pph23_percentage DECIMAL(20, 5),
  vat_percentage DECIMAL(20, 5),
  discount_amount DECIMAL(20, 5),
  discount_percentage DECIMAL(20, 5),
  discount_percentage_amount DECIMAL(20, 5),
  discount_final DECIMAL(20, 5),
  discount_type TEXT,
  dp_percentage DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  total_amount_products DECIMAL(20, 5),
  total_dp_products DECIMAL(20, 5),
  total_qty DECIMAL(20, 5),
  total_discount DECIMAL(20, 5),
  total_pph23 DECIMAL(20, 5),
  total_vat DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_invoice_dps_customer_id ON invoice_dps(customer_id);

CREATE INDEX idx_invoice_dps_currency_id ON invoice_dps(currency_id);

CREATE INDEX idx_invoice_dps_payment_term_id ON invoice_dps(payment_term_id);

CREATE INDEX idx_invoice_dps_vat_id ON invoice_dps(vat_id);

CREATE INDEX idx_invoice_dps_pph23_id ON invoice_dps(pph23_id);

CREATE INDEX idx_invoice_dps_branch_id ON invoice_dps(branch_id);

COMMIT;