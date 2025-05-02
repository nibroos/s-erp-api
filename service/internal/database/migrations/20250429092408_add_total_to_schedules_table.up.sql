ALTER TABLE
  schedules
ADD
  COLUMN total_task_step_4_done INT DEFAULT 0;

ALTER TABLE
  schedules
ADD
  COLUMN total_all_tasks_done INT DEFAULT 0;

ALTER TABLE
  schedules
ADD
  COLUMN total_tasks INT DEFAULT 0;