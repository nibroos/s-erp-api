DROP INDEX IF EXISTS idx_inv_dts_ref_ro_dt_id;

ALTER TABLE
  inv_dts DROP COLUMN ref_ro_dt_id;