CREATE INDEX idx_inv_dts_inventory_id ON inv_dts(inventory_id);

CREATE INDEX idx_inv_dts_item_unit_id ON inv_dts(item_unit_id);

CREATE INDEX idx_inv_dts_vat_id ON inv_dts(vat_id);

CREATE INDEX idx_inv_dts_pph23_id ON inv_dts(pph23_id);

CREATE INDEX idx_inv_dts_ref_so_dt_id ON inv_dts(ref_so_dt_id);

CREATE INDEX idx_inv_dts_ref_so_dt_bom_id ON inv_dts(ref_so_dt_bom_id);

CREATE INDEX idx_inv_dts_ref_po_dt_id ON inv_dts(ref_po_dt_id);

CREATE INDEX idx_inv_dts_ref_po_dt_bom_id ON inv_dts(ref_po_dt_bom_id);

CREATE INDEX idx_inv_dts_ref_inv_dt_id ON inv_dts(ref_inv_dt_id);

CREATE INDEX idx_inv_dts_ref_product_id ON inv_dts(ref_product_id);

CREATE INDEX idx_inv_dts_item_id ON inv_dts(item_id);

CREATE INDEX idx_inventories_customer_id ON inventories(customer_id);

CREATE INDEX idx_inventories_io_type_id ON inventories(io_type_id);

CREATE INDEX idx_inventories_currency_id ON inventories(currency_id);

CREATE INDEX idx_inventories_payment_term_id ON inventories(payment_term_id);

CREATE INDEX idx_inventories_warehouse_id ON inventories(warehouse_id);

CREATE INDEX idx_inventories_vat_id ON inventories(vat_id);

CREATE INDEX idx_inventories_pph23_id ON inventories(pph23_id);

CREATE INDEX idx_inventories_branch_id ON inventories(branch_id);