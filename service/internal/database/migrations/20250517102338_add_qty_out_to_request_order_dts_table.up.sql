ALTER TABLE
  request_order_dts
ADD
  COLUMN qty_out DECIMAL(20, 5) DEFAULT 0;

UPDATE
  request_order_dts
SET
  qty_out = 0
WHERE
  qty_out IS NULL;