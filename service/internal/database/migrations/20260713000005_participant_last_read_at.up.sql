-- When a participant last read the conversation (drives "seen at <time>").
ALTER TABLE conversation_participants ADD COLUMN IF NOT EXISTS last_read_at timestamp with time zone;
