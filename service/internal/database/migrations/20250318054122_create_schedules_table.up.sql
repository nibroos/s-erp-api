BEGIN;

CREATE TABLE IF NOT EXISTS schedules (
  id SERIAL PRIMARY KEY,
  assignee_id INT REFERENCES users(id) ON DELETE RESTRICT,
  -- sales_order_id INT REFERENCES sales_orders(id) ON DELETE RESTRICT,
  sales_order_id INT,
  uuid TEXT,
  steps_id INT,
  module_type TEXT,
  title TEXT,
  remark TEXT,
  status TEXT DEFAULT 'WAITING',
  start_at date,
  end_at date,
  color TEXT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN schedules.status IS 'WAITING, PROCESS, FINISHED, CANCELED';

COMMENT ON COLUMN schedules.module_type IS 'sales_orders, feedbacks';

CREATE INDEX idx_schedules_assignee_id ON schedules(assignee_id);

CREATE INDEX idx_schedules_sales_order_id ON schedules(sales_order_id);

COMMIT;