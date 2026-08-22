-- Optional per-attachment caption/description.
ALTER TABLE message_attachments ADD COLUMN IF NOT EXISTS description TEXT;
