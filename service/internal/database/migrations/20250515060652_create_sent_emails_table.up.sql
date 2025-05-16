CREATE TABLE IF NOT EXISTS sent_emails (
  id SERIAL PRIMARY KEY,
  ref_id INT,
  sender_id BIGINT REFERENCES users(id) ON DELETE RESTRICT,
  ref_type TEXT,
  from_email TEXT,
  to_email TEXT,
  subject TEXT,
  remark TEXT,
  error_message TEXT,
  log_json JSONB,
  status TEXT DEFAULT 'PROCESS',
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN sent_emails.status IS 'PROCESS, SUCCESS, FAILED';

COMMENT ON COLUMN sent_emails.ref_type IS 'tickets, invoice_maintenances';