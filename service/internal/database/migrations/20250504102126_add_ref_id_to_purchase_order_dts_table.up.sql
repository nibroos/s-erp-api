ALTER TABLE
  purchase_order_dts
ADD
  COLUMN ref_so_dt_id BIGINT REFERENCES so_dts(id) ON DELETE RESTRICT;

ALTER TABLE
  purchase_order_dts
ADD
  COLUMN ref_so_dt_bom_id BIGINT REFERENCES so_dt_boms(id) ON DELETE RESTRICT;

ALTER TABLE
  purchase_order_dts
ADD
  COLUMN ref_product_id BIGINT REFERENCES products(id) ON DELETE RESTRICT;

ALTER TABLE
  purchase_order_dts
ADD
  COLUMN vat_perc DECIMAL(20, 5) DEFAULT 0;

ALTER TABLE
  purchase_order_dts
ADD
  COLUMN vat_perc_am DECIMAL(20, 5) DEFAULT 0;

ALTER TABLE
  purchase_order_dts
ADD
  COLUMN pph23_perc DECIMAL(20, 5) DEFAULT 0;

ALTER TABLE
  purchase_order_dts
ADD
  COLUMN pph23_perc_am DECIMAL(20, 5) DEFAULT 0;