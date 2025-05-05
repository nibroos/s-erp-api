ALTER TABLE
  purchase_order_dts
ADD
  COLUMN qty_in DECIMAL(20, 5) DEFAULT 0;

UPDATE
  purchase_order_dts
SET
  qty_in = 0
WHERE
  qty_in IS NULL;