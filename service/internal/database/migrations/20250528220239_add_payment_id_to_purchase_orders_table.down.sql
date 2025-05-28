DROP INDEX idx_purchase_orders_payment_id;

ALTER TABLE
  purchase_orders DROP COLUMN payment_id;