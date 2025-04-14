CREATE TABLE IF NOT EXISTS mix_values (
  id SERIAL PRIMARY KEY,
  group_id INT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
  parent_id INT,
  name VARCHAR(500) NOT NULL,
  description TEXT,
  date_at DATE,
  time_at TIME,
  remark TEXT,
  num DECIMAL(20, 5),
  order_item DECIMAL(20, 5),
  color TEXT,
  status INT,
  options_json JSONB,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_mix_values_group_id ON mix_values (group_id);

CREATE INDEX idx_mix_values_parent_id ON mix_values (parent_id);

CREATE INDEX idx_mix_values_name ON mix_values (name);