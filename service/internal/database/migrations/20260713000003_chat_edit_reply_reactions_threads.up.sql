-- Message editing, replies, and message kind (text | system).
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edited_at timestamp with time zone;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS reply_to_id BIGINT REFERENCES messages(id) ON DELETE SET NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'text';
CREATE INDEX IF NOT EXISTS idx_messages_reply_to_id ON messages (reply_to_id);

-- Threads: a thread is a conversation branched from a message in a parent
-- conversation. parent_id links it to the parent; root_message_id is the
-- message it was started from.
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS parent_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS root_message_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_conversations_parent_id ON conversations (parent_id);
CREATE INDEX IF NOT EXISTS idx_conversations_root_message_id ON conversations (root_message_id);

-- Emoji reactions on messages.
CREATE TABLE IF NOT EXISTS message_reactions (
  id BIGSERIAL PRIMARY KEY,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  emoji VARCHAR(16) NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_message_reactions ON message_reactions (message_id, user_id, emoji);
CREATE INDEX IF NOT EXISTS idx_message_reactions_message ON message_reactions (message_id);

-- @mentions inside a message.
CREATE TABLE IF NOT EXISTS message_mentions (
  id BIGSERIAL PRIMARY KEY,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_message_mentions ON message_mentions (message_id, user_id);
CREATE INDEX IF NOT EXISTS idx_message_mentions_message ON message_mentions (message_id);
