BEGIN;

CREATE TABLE IF NOT EXISTS purchase_orders (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  purchase_type_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  currency_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  vat_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  payment_term_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  shipping_term_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  pph23_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  branch_id INT REFERENCES branches(id) ON DELETE RESTRICT,
  po_no TEXT,
  po_date timestamp with time zone,
  delivery_date timestamp with time zone,
  shipping_destination TEXT,
  remark TEXT,
  exchange_rate DECIMAL(20, 5),
  discount_percentage DECIMAL(20, 5),
  discount_amount DECIMAL(20, 5),
  discount_percentage_amount DECIMAL(20, 5),
  discount_final_header DECIMAL(20, 5),
  discount_amount_product DECIMAL(20, 5),
  discount_type TEXT,
  pph23_percentage DECIMAL(20, 5),
  vat_percentage DECIMAL(20, 5),
  total_amount_products DECIMAL(20, 5),
  subtotal DECIMAL(20, 5),
  total_qty DECIMAL(20, 5),
  total_discount DECIMAL(20, 5),
  total_pph23 DECIMAL(20, 5),
  total_vat DECIMAL(20, 5),
  grand_total DECIMAL(20, 5),
  status TEXT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_purchase_orders_customer_id ON purchase_orders(customer_id);
CREATE INDEX idx_purchase_orders_purchase_type_id ON purchase_orders(purchase_type_id);
CREATE INDEX idx_purchase_orders_currency_id ON purchase_orders(currency_id);
CREATE INDEX idx_purchase_orders_vat_id ON purchase_orders(vat_id);
CREATE INDEX idx_purchase_orders_payment_term_id ON purchase_orders(payment_term_id);
CREATE INDEX idx_purchase_orders_shipping_term_id ON purchase_orders(shipping_term_id);
CREATE INDEX idx_purchase_orders_pph23_id ON purchase_orders(pph23_id);
CREATE INDEX idx_purchase_orders_branch_id ON purchase_orders(branch_id);

COMMIT;
