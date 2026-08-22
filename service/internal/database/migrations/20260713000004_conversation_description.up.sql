-- Optional description / topic for groups and threads.
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS description TEXT;
