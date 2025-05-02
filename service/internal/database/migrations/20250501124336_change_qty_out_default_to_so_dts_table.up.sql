UPDATE
  so_dts
SET
  qty_out = 0
WHERE
  qty_out IS NULL;

ALTER TABLE
  so_dts
ALTER COLUMN
  qty_out
SET
  NOT NULL;

ALTER TABLE
  so_dts
ALTER COLUMN
  qty_out
SET
  DEFAULT 0;

UPDATE
  so_dts
SET
  qty_out = 0
WHERE
  qty_out IS NULL;