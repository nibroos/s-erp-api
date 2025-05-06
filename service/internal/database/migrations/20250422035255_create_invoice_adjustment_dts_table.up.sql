CREATE TABLE IF NOT EXISTS invoice_adjustment_dts (
  id SERIAL PRIMARY KEY,
  invoice_uuid TEXT,
  invoice_adjustment_id INT REFERENCES invoice_adjustments(id) ON DELETE RESTRICT,
  ref_id INT,
  ref_type TEXT,
  ref_json JSONB,
  invoice_no TEXT,
  invoice_date date,
  invoice_amount DECIMAL(20, 5),
  total_adjustment DECIMAL(20, 5),
  balance_amount DECIMAL(20, 5),
  adjustment_amount DECIMAL(20, 5),
  admin_bank DECIMAL(20, 5),
  total_amount DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_invoice_adjustment_dts_invoice_adjustment_id ON invoice_adjustment_dts(invoice_adjustment_id);

CREATE INDEX idx_invoice_adjustment_dts_ref_id ON invoice_adjustment_dts(ref_id);