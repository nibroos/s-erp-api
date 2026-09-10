DELETE FROM users WHERE email = 'assistant@erp.ai' AND is_ai = true;
ALTER TABLE users DROP COLUMN IF EXISTS is_ai;
