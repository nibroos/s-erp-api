CREATE TABLE IF NOT EXISTS invoice_adjustments (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  bank_id INT REFERENCES bank_informations(id) ON DELETE RESTRICT,
  invoice_no TEXT,
  payment_date date,
  payment_amount DECIMAL(20, 5),
  exchange_rate DECIMAL(20, 5),
  reference TEXT,
  ref_start_date date,
  ref_end_date date,
  remark TEXT,
  rev_no INT,
  total_invoice DECIMAL(20, 5),
  total_adjustment DECIMAL(20, 5),
  total_balance DECIMAL(20, 5),
  total_admin_bank DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_invoice_adjustments_customer_id ON invoice_adjustments(customer_id);

CREATE INDEX idx_invoice_adjustments_currency_id ON invoice_adjustments(currency_id);

CREATE INDEX idx_invoice_adjustments_bank_id ON invoice_adjustments(bank_id);

CREATE INDEX idx_invoice_adjustments_branch_id ON invoice_adjustments(branch_id);