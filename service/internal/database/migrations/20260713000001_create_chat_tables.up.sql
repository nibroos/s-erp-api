-- Conversations: a chat thread. type = 'direct' (1-1) or 'group'.
-- direct_key holds the deterministic key "minUserId:maxUserId" for direct
-- conversations so a 1-1 thread between two users can be looked up / created
-- atomically and can never be duplicated.
CREATE TABLE IF NOT EXISTS conversations (
  id BIGSERIAL PRIMARY KEY,
  type VARCHAR(20) NOT NULL DEFAULT 'direct',
  title VARCHAR(255),
  direct_key VARCHAR(64),
  last_message_id BIGINT,
  last_message_at timestamp with time zone,
  created_by_id BIGINT,
  updated_by_id BIGINT,
  deleted_by_id BIGINT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_conversations_direct_key ON conversations (direct_key) WHERE direct_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_conversations_last_message_at ON conversations (last_message_at DESC);

-- Participants of a conversation. last_read_message_id drives unread counts.
CREATE TABLE IF NOT EXISTS conversation_participants (
  id BIGSERIAL PRIMARY KEY,
  conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  last_read_message_id BIGINT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_participants_conversation_user ON conversation_participants (conversation_id, user_id);
CREATE INDEX IF NOT EXISTS idx_participants_user_id ON conversation_participants (user_id);

-- Messages within a conversation.
CREATE TABLE IF NOT EXISTS messages (
  id BIGSERIAL PRIMARY KEY,
  conversation_id BIGINT NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  content TEXT NOT NULL,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  deleted_at timestamp with time zone
);

-- Keyset pagination + fetching a conversation history quickly.
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id_id ON messages (conversation_id, id DESC);
