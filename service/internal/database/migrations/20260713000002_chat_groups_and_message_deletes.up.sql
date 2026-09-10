-- Participant role within a conversation: 'owner' | 'admin' | 'member'.
ALTER TABLE conversation_participants
  ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'member';

-- Message deletion state.
-- deleted_for_all: soft "delete for everyone" (row kept, content blanked, shown
-- as a "message deleted" placeholder). Distinct from deleted_at (hard soft-delete).
ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS deleted_for_all BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE messages
  ADD COLUMN IF NOT EXISTS deleted_by_id BIGINT;

-- Per-user "delete for me": hides a message only for the given user.
CREATE TABLE IF NOT EXISTS message_hides (
  id BIGSERIAL PRIMARY KEY,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_message_hides_message_user ON message_hides (message_id, user_id);
CREATE INDEX IF NOT EXISTS idx_message_hides_user_id ON message_hides (user_id);

-- Existing direct conversations: make both participants 'owner' is wrong; keep
-- them as 'member' (default). Nothing else needed.
