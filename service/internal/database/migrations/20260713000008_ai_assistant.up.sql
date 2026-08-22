-- Flag distinguishing AI assistant accounts from human users.
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_ai BOOLEAN NOT NULL DEFAULT false;

-- Seed the built-in ERP AI assistant. It is a normal user row (so it reuses the
-- whole chat/participant/message infrastructure) but flagged is_ai. Login is
-- impossible: the password hash is a placeholder.
INSERT INTO users (name, username, email, password, status, is_ai, created_at, updated_at)
VALUES ('ERP Assistant', 'erp-assistant', 'assistant@erp.ai', 'x', 1, true, now(), now())
ON CONFLICT (email) DO UPDATE SET is_ai = true;
