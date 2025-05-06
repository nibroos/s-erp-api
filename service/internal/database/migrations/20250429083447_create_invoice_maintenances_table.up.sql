CREATE TABLE IF NOT EXISTS invoice_maintenances (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  payment_term_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  bank_id INT REFERENCES bank_informations(id) ON DELETE RESTRICT,
  invoice_no TEXT,
  invoice_date date,
  exchange_rate DECIMAL(20, 5),
  remark TEXT,
  rev_no INT,
  status TEXT,
  approved_status TEXT,
  pph23_percentage DECIMAL(20, 5),
  vat_percentage DECIMAL(20, 5),
  discount_amount DECIMAL(20, 5),
  discount_percentage DECIMAL(20, 5),
  discount_percentage_amount DECIMAL(20, 5),
  discount_final DECIMAL(20, 5),
  discount_type TEXT,
  subtotal DECIMAL(20, 5),
  total_amount_products DECIMAL(20, 5),
  total_dp_products DECIMAL(20, 5),
  total_balance_products DECIMAL(20, 5),
  total_qty DECIMAL(20, 5),
  total_discount DECIMAL(20, 5),
  total_pph23 DECIMAL(20, 5),
  total_vat DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  approved_by_id INT, 
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone,
  total_adjustment DECIMAL(20, 5),
  history_total_adjustment DECIMAL(20, 5),
  history_status TEXT
);

CREATE INDEX idx_invoice_maintenances_customer_id ON invoice_maintenances(customer_id);

CREATE INDEX idx_invoice_maintenances_currency_id ON invoice_maintenances(currency_id);

CREATE INDEX idx_invoice_maintenances_payment_term_id ON invoice_maintenances(payment_term_id);

CREATE INDEX idx_invoice_maintenances_vat_id ON invoice_maintenances(vat_id);

CREATE INDEX idx_invoice_maintenances_pph23_id ON invoice_maintenances(pph23_id);

CREATE INDEX idx_invoice_maintenances_branch_id ON invoice_maintenances(branch_id);

CREATE INDEX idx_invoice_maintenances_bank_id ON invoice_maintenances(bank_id);