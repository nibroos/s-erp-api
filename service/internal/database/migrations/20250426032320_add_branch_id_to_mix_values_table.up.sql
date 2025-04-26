ALTER TABLE
  mix_values
ADD
  COLUMN branch_id BIGINT REFERENCES branches(id) ON DELETE RESTRICT;

ALTER TABLE
  mix_values
ADD
  COLUMN is_main INT DEFAULT 0;

CREATE INDEX idx_mix_values_branch_id ON mix_values(branch_id);