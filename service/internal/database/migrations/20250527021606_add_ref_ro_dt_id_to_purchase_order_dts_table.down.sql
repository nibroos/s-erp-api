DROP INDEX IF EXISTS idx_purchase_order_dts_ref_ro_dt_id;

ALTER TABLE
  purchase_order_dts DROP COLUMN ref_ro_dt_id;