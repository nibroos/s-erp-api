ALTER TABLE
  so_dts
ADD
  COLUMN qty_po DECIMAL(20, 5) DEFAULT 0;

ALTER TABLE
  so_dt_boms
ADD
  COLUMN qty_po DECIMAL(20, 5) DEFAULT 0;

UPDATE
  so_dts
SET
  qty_po = 0
WHERE
  qty_po IS NULL;

UPDATE
  so_dt_boms
SET
  qty_po = 0
WHERE
  qty_po IS NULL;