CREATE TABLE IF NOT EXISTS pic_emails (
  id BIGSERIAL PRIMARY KEY,
  customer_id BIGINT REFERENCES customers(id) ON DELETE RESTRICT,
  name TEXT,
  is_main INT,
  created_by_id BIGINT,
  updated_by_id BIGINT,
  deleted_by_id BIGINT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);