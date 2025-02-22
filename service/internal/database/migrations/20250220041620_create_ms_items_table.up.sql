BEGIN;

CREATE TABLE IF NOT EXISTS ms_items (
  id SERIAL PRIMARY KEY,
  item_sub_group_id INT NOT NULL,
  item_unit_id INT,
  code TEXT,
  name VARCHAR(255) NOT NULL,
  specification TEXT,
  description TEXT,
  tpb_code TEXT,
  minimum_stock DECIMAL(18, 5),
  is_all_branch INT,
  status INT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMIT;