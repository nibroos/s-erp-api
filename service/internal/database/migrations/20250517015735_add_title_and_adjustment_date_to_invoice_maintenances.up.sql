ALTER TABLE invoice_adjustments ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE invoice_adjustments ADD COLUMN IF NOT EXISTS adjustment_date date;