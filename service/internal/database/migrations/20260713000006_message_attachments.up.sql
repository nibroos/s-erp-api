-- Attachments (documents, images, video, voice) stored in MinIO. Only the
-- object key is persisted; a time-limited presigned URL is generated on read.
CREATE TABLE IF NOT EXISTS message_attachments (
  id BIGSERIAL PRIMARY KEY,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  kind VARCHAR(20) NOT NULL,          -- image | video | voice | document
  object_key VARCHAR(512) NOT NULL,
  file_name VARCHAR(255),
  mime_type VARCHAR(128),
  size_bytes BIGINT,
  duration_seconds INTEGER,
  width INTEGER,
  height INTEGER,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_message_attachments_message ON message_attachments (message_id);
