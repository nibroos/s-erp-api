BEGIN;

CREATE TABLE IF NOT EXISTS purchase_order_dt_boms (
  id SERIAL PRIMARY KEY,
  po_id INT REFERENCES purchase_orders(id) ON DELETE RESTRICT,
  po_dt_id INT REFERENCES purchase_order_dts(id) ON DELETE RESTRICT,
  bom_id INT REFERENCES boms(id) ON DELETE RESTRICT,
  product_id INT REFERENCES products(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  product_json JSONB,
  gen_code TEXT,
  remark TEXT,
  qty DECIMAL(20, 5),
  price DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_purchase_order_dt_boms_po_id ON purchase_order_dt_boms(po_id);
CREATE INDEX idx_purchase_order_dt_boms_po_dt_id ON purchase_order_dt_boms(po_dt_id);
CREATE INDEX idx_purchase_order_dt_boms_bom_id ON purchase_order_dt_boms(bom_id);
CREATE INDEX idx_purchase_order_dt_boms_product_id ON purchase_order_dt_boms(product_id);
CREATE INDEX idx_purchase_order_dt_boms_item_unit_id ON purchase_order_dt_boms(item_unit_id);

COMMIT;
