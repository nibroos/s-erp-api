BEGIN;

CREATE TABLE IF NOT EXISTS item_units (
  id SERIAL PRIMARY KEY,
  unit_id INT NOT NULL REFERENCES mix_values (id) ON DELETE RESTRICT,
  ms_item_id INT NOT NULL REFERENCES ms_items (id) ON DELETE RESTRICT,
  conversion DECIMAL(20, 5),
  price_sell DECIMAL(20, 5),
  price_buy DECIMAL(20, 5),
  status INT,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_item_units_unit_id ON item_units(unit_id);

CREATE INDEX idx_item_units_ms_item_id ON item_units(ms_item_id);

ALTER TABLE
  ms_items
ADD
  CONSTRAINT fk_ms_items_item_unit_id FOREIGN KEY (item_unit_id) REFERENCES item_units (id) ON DELETE RESTRICT;

COMMIT;