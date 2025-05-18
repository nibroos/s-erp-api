ALTER TABLE
  inv_dts
ADD
  COLUMN ref_ro_dt_id BIGINT REFERENCES request_order_dts(id) ON DELETE RESTRICT;

-- add index
CREATE INDEX idx_inv_dts_ref_ro_dt_id ON inv_dts(ref_ro_dt_id);