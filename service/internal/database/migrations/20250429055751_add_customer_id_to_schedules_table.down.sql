DROP INDEX IF EXISTS idx_schedules_customer_id;

-- Then drop the columns in reverse order
ALTER TABLE
  schedules DROP COLUMN IF EXISTS customer_id;