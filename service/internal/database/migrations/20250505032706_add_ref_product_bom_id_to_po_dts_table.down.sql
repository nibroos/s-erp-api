DROP INDEX idx_purchase_order_dts_ref_product_bom_id;

DROP INDEX idx_inv_dts_ref_product_bom_id;

ALTER TABLE
  purchase_order_dts DROP COLUMN ref_product_bom_id;

ALTER TABLE
  inv_dts DROP COLUMN ref_product_bom_id;