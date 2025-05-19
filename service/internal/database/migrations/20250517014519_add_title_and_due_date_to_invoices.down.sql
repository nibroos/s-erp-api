ALTER TABLE invoice_dps DROP COLUMN IF EXISTS title;
ALTER TABLE invoice_dps DROP COLUMN IF EXISTS due_date;

ALTER TABLE sales_invoices DROP COLUMN IF EXISTS title;
ALTER TABLE sales_invoices DROP COLUMN IF EXISTS due_date;

ALTER TABLE invoice_maintenances DROP COLUMN IF EXISTS title;
ALTER TABLE invoice_maintenances DROP COLUMN IF EXISTS due_date;
