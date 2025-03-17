BEGIN;

CREATE TABLE IF NOT EXISTS purchase_order_dts (
  id SERIAL PRIMARY KEY,
  po_id INT REFERENCES purchase_orders(id) ON DELETE RESTRICT,
  item_unit_id INT REFERENCES item_units(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  ref_id INT,
  product_id INT REFERENCES products(id) ON DELETE RESTRICT,
  product_type TEXT,
  product_json JSONB,
  ref_type TEXT,
  ref_json JSONB,
  gen_code TEXT,
  remark TEXT,
  need_qty DECIMAL(20, 5),
  qty DECIMAL(20, 5),
  price DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  discount_amount DECIMAL(20, 5),
  discount_percentage DECIMAL(20, 5),
  discount_percentage_num DECIMAL(20, 5),
  discount_percentage_amount DECIMAL(20, 5),
  discount_final DECIMAL(20, 5),
  discount_type TEXT,
  vat_percentage DECIMAL(20, 5),
  vat_percentage_amount DECIMAL(20, 5),
  total_amount DECIMAL(20, 5),
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_purchase_order_dts_po_id ON purchase_order_dts(po_id);
CREATE INDEX idx_purchase_order_dts_item_unit_id ON purchase_order_dts(item_unit_id);
CREATE INDEX idx_purchase_order_dts_vat_id ON purchase_order_dts(vat_id);
CREATE INDEX idx_purchase_order_dts_ref_id ON purchase_order_dts(ref_id);
CREATE INDEX idx_purchase_order_dts_product_id ON purchase_order_dts(product_id);

COMMIT;
