ALTER TABLE
  schedules
ADD
  COLUMN customer_id BIGINT REFERENCES customers(id) ON DELETE RESTRICT;

CREATE INDEX idx_schedules_customer_id ON schedules(customer_id);