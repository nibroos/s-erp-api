CREATE TABLE IF NOT EXISTS customer_contracts (
  id BIGSERIAL PRIMARY KEY,
  customer_id BIGINT REFERENCES customers(id) ON DELETE RESTRICT,
  payment_type_id BIGINT REFERENCES mix_values(id) ON DELETE RESTRICT,
  product_id BIGINT REFERENCES products(id) ON DELETE RESTRICT,
  price DECIMAL(20, 5),
  agree_at date,
  due_at date,
  qty DECIMAL(20, 5),
  installation_at date,
  warranty_at date,
  remark TEXT,
  created_by_id BIGINT,
  updated_by_id BIGINT,
  deleted_by_id BIGINT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);