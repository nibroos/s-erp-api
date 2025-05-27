ALTER TABLE
  purchase_order_dts
ADD
  COLUMN ref_ro_dt_id BIGINT REFERENCES request_order_dts(id) ON DELETE RESTRICT;

-- add index
CREATE INDEX idx_purchase_order_dts_ref_ro_dt_id ON purchase_order_dts(ref_ro_dt_id);