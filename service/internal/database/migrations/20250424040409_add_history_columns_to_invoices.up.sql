ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS history_total_adjustment DECIMAL(20, 2);
ALTER TABLE invoice_dps ADD COLUMN IF NOT EXISTS history_total_adjustment DECIMAL(20, 2);
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS history_status TEXT;
ALTER TABLE invoice_dps ADD COLUMN IF NOT EXISTS history_status TEXT;