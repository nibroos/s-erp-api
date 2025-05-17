ALTER TABLE invoice_dps ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE invoice_dps ADD COLUMN IF NOT EXISTS due_date date;

ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE sales_invoices ADD COLUMN IF NOT EXISTS due_date date;

ALTER TABLE invoice_maintenances ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE invoice_maintenances ADD COLUMN IF NOT EXISTS due_date date;