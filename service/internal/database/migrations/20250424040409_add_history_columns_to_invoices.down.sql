ALTER TABLE sales_invoices DROP COLUMN IF EXISTS history_total_adjustment;
ALTER TABLE invoice_dps DROP COLUMN IF EXISTS history_total_adjustment;
ALTER TABLE sales_invoices DROP COLUMN IF EXISTS history_status;
ALTER TABLE invoice_dps DROP COLUMN IF EXISTS history_status;