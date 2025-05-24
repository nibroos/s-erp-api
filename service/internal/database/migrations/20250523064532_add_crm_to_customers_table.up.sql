ALTER TABLE
  customers
ADD
  COLUMN remark TEXT;

ALTER TABLE
  customers
ADD
  COLUMN owner_name TEXT;

ALTER TABLE
  customers
ADD
  COLUMN owner_phone TEXT;

ALTER TABLE
  customers
ADD
  COLUMN owner_email TEXT;

ALTER TABLE
  customers
ADD
  COLUMN category_type_id BIGINT,
ADD
  CONSTRAINT fk_category_type_id FOREIGN KEY (category_type_id) REFERENCES mix_values(id) ON DELETE
SET
  NULL ON UPDATE CASCADE;

ALTER TABLE
  customers
ADD
  COLUMN contract_date DATE;

ALTER TABLE
  customers
ADD
  COLUMN is_contract INT8;

ALTER TABLE
  customers
ADD
  COLUMN pic_name TEXT;

ALTER TABLE
  customers
ADD
  COLUMN pic_phone TEXT;

CREATE INDEX idx_customers_category_type_id ON customers(category_type_id);