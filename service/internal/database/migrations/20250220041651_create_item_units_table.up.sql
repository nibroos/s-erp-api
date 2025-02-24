BEGIN;

CREATE TABLE IF NOT EXISTS item_units (
  id SERIAL PRIMARY KEY,
  unit_id INT NOT NULL,
  ms_item_id INT NOT NULL,
  conversion DECIMAL(18, 5),
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