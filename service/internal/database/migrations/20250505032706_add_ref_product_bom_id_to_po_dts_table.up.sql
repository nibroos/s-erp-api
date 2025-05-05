ALTER TABLE
  purchase_order_dts
ADD
  COLUMN ref_product_bom_id BIGINT NULL,
ADD
  CONSTRAINT fk_ref_product_bom_id FOREIGN KEY (ref_product_bom_id) REFERENCES boms(id) ON DELETE
SET
  NULL ON UPDATE CASCADE;

ALTER TABLE
  inv_dts
ADD
  COLUMN ref_product_bom_id BIGINT NULL,
ADD
  CONSTRAINT fk_ref_product_bom_id FOREIGN KEY (ref_product_bom_id) REFERENCES boms(id) ON DELETE
SET
  NULL ON UPDATE CASCADE;

CREATE INDEX idx_purchase_order_dts_ref_product_bom_id ON purchase_order_dts(ref_product_bom_id);

CREATE INDEX idx_inv_dts_ref_product_bom_id ON inv_dts(ref_product_bom_id);