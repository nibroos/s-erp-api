DROP TABLE IF EXISTS message_mentions;
DROP TABLE IF EXISTS message_reactions;
ALTER TABLE conversations DROP COLUMN IF EXISTS root_message_id;
ALTER TABLE conversations DROP COLUMN IF EXISTS parent_id;
ALTER TABLE messages DROP COLUMN IF EXISTS type;
ALTER TABLE messages DROP COLUMN IF EXISTS reply_to_id;
ALTER TABLE messages DROP COLUMN IF EXISTS edited_at;
