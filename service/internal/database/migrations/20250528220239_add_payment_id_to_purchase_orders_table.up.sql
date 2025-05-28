ALTER TABLE
  purchase_orders
ADD
  COLUMN payment_id BIGINT REFERENCES bank_informations(id) ON DELETE RESTRICT;

CREATE INDEX idx_purchase_orders_payment_id ON purchase_orders(payment_id);