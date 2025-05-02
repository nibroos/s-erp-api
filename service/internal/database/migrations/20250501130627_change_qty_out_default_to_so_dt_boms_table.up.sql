UPDATE
  so_dt_boms
SET
  qty_out = 0
WHERE
  qty_out IS NULL;

ALTER TABLE
  so_dt_boms
ALTER COLUMN
  qty_out
SET
  NOT NULL;

ALTER TABLE
  so_dt_boms
ALTER COLUMN
  qty_out
SET
  DEFAULT 0;

UPDATE
  so_dt_boms
SET
  qty_out = 0
WHERE
  qty_out IS NULL;