BEGIN;

CREATE TABLE IF NOT EXISTS pools (
  id SERIAL PRIMARY KEY,
  group1_id INT REFERENCES groups(id) ON DELETE RESTRICT,
  group2_id INT REFERENCES groups(id) ON DELETE RESTRICT,
  mv1_id INT,
  mv2_id INT,
  description VARCHAR(255),
  options_json JSONB,
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

CREATE INDEX idx_pools_group1_id ON pools (group1_id);

CREATE INDEX idx_pools_group2_id ON pools (group2_id);

CREATE INDEX idx_pools_mv1_id ON pools (mv1_id);

CREATE INDEX idx_pools_mv2_id ON pools (mv2_id);

COMMIT;