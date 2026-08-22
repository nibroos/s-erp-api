DROP TABLE IF EXISTS message_hides;
ALTER TABLE messages DROP COLUMN IF EXISTS deleted_for_all;
ALTER TABLE messages DROP COLUMN IF EXISTS deleted_by_id;
ALTER TABLE conversation_participants DROP COLUMN IF EXISTS role;
