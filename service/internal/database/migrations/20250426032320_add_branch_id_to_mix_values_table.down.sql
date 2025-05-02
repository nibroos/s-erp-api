DROP INDEX IF EXISTS idx_mix_values_branch_id;

ALTER TABLE
  mix_values DROP COLUMN IF EXISTS is_main;

ALTER TABLE
  mix_values DROP COLUMN IF EXISTS branch_id;