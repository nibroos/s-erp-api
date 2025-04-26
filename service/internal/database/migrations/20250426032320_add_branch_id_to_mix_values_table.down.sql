-- First drop the index
DROP INDEX IF EXISTS idx_mix_values_branch_id;

-- Then drop the columns in reverse order
ALTER TABLE
  mix_values DROP COLUMN IF EXISTS is_main;

ALTER TABLE
  mix_values DROP COLUMN IF EXISTS branch_id;