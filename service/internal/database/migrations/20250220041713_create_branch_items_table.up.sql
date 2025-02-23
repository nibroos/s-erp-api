BEGIN;

CREATE TABLE IF NOT EXISTS branch_items (
  id SERIAL PRIMARY KEY,
  branch_id INT NOT NULL,
  item_unit_id INT,
  name VARCHAR(255) NOT NULL,
  specification TEXT,
  description TEXT,
  tpb_code TEXT,
  minimum_stock DECIMAL(18, 5),
  price_sell DECIMAL(18, 5),
  price_buy DECIMAL(18, 5),
  status INT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMIT;