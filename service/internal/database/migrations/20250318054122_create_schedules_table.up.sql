BEGIN;

CREATE TABLE IF NOT EXISTS schedules (
  id SERIAL PRIMARY KEY,
  customer_id INT REFERENCES customers(id) ON DELETE RESTRICT,
  assignee_id INT REFERENCES users(id) ON DELETE RESTRICT,
  sales_order_id INT REFERENCES sales_orders(id) ON DELETE RESTRICT,
  schedule_type_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  schedule_no TEXT,
  title TEXT,
  remark TEXT,
  status TEXT DEFAULT 'WAITING',
  start_at date,
  end_at date,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMIT;