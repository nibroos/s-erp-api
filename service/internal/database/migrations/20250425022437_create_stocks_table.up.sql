CREATE TABLE IF NOT EXISTS stocks (
  id SERIAL PRIMARY KEY,
  warehouse_id BIGINT REFERENCES mix_values(id) ON DELETE RESTRICT,
  item_id BIGINT REFERENCES products(id) ON DELETE RESTRICT,
  branch_id BIGINT REFERENCES branches(id) ON DELETE RESTRICT,
  qty DECIMAL(20, 5),
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);