-- First drop any foreign key constraints
ALTER TABLE
  customers DROP CONSTRAINT IF EXISTS fk_customers_category_type;

-- Then drop any indexes
DROP INDEX IF EXISTS idx_customers_category_type_id;

-- Finally drop the column
ALTER TABLE
  customers DROP COLUMN category_type_id;

alter TABLE
  customers DROP COLUMN remark;

ALTER TABLE
  customers DROP COLUMN owner_name;

ALTER TABLE
  customers DROP COLUMN owner_phone;

ALTER TABLE
  customers DROP COLUMN owner_email;

ALTER TABLE
  customers DROP COLUMN contract_date;

ALTER TABLE
  customers DROP COLUMN is_contract;

ALTER TABLE
  customers DROP COLUMN pic_name;

ALTER TABLE
  customers DROP COLUMN pic_phone;