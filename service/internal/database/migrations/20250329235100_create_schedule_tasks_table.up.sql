BEGIN;

CREATE TYPE schedule_task_entity_type AS ENUM ('steps', 'tasks', 'comments');

CREATE TABLE IF NOT EXISTS schedule_tasks (
  id SERIAL PRIMARY KEY,
  schedule_id INT REFERENCES schedules(id) ON DELETE RESTRICT,
  assignee_id INT REFERENCES users(id) ON DELETE RESTRICT,
  parent_id INT,
  entity_id INT REFERENCES mix_values(id) ON DELETE RESTRICT,
  entity_type schedule_task_entity_type,
  uuid TEXT,
  parent_uuid TEXT,
  title TEXT,
  remark TEXT,
  order_item INT,
  color TEXT,
  is_checked INT DEFAULT 0,
  locations JSONB DEFAULT '{}',
  start_at timestamp,
  end_at timestamp,
  options_json JSONB DEFAULT '{}',
  created_by_id INT,
  updated_by_id INT,
  deleted_by_id INT,
  created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
  updated_at timestamp with time zone,
  deleted_at timestamp with time zone
);

COMMENT ON COLUMN schedule_tasks.entity_type IS 'steps, tasks, comments';

CREATE INDEX idx_schedule_tasks_assignee_id ON schedule_tasks(assignee_id);

CREATE INDEX idx_schedule_tasks_schedule_id ON schedule_tasks(schedule_id);

CREATE INDEX idx_schedule_tasks_parent_id ON schedule_tasks(parent_id);

CREATE INDEX idx_schedule_tasks_entity_type ON schedule_tasks(entity_type);

COMMIT;