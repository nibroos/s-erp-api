-- Dropping the schema takes every view with it; the role must lose its grants
-- before it can be dropped.
DROP SCHEMA IF EXISTS ai CASCADE;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ai_readonly') THEN
    DROP ROLE ai_readonly;
  END IF;
END
$$;
