ALTER TABLE
  purchase_orders
ADD
  COLUMN rev_no INT DEFAULT 0;

ALTER TABLE
  purchase_orders
ADD
  COLUMN po_no_ori TEXT;