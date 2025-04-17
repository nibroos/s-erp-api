BEGIN;

ALTER TABLE invoice_dps
ADD COLUMN bank_id INTEGER REFERENCES bank_informations(id);

CREATE INDEX idx_invoice_dps_bank_id ON invoice_dps(bank_id);

ALTER TABLE sales_invoices
ADD COLUMN bank_id INTEGER REFERENCES bank_informations(id);

CREATE INDEX idx_sales_invoices_bank_id ON sales_invoices(bank_id);

COMMIT;
