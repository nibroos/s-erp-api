BEGIN;

INSERT INTO
  mix_values (
    group_id,
    name,
    description,
    status,
    options_json,
    created_at,
    updated_at
  )
VALUES
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 1', 'Task 1 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 2', 'Task 2 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 3', 'Task 3 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 4', 'Task 4 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 5', 'Task 5 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 6', 'Task 6 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 7', 'Task 7 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 8', 'Task 8 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 9', 'Task 9 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 10', 'Task 10 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 11', 'Task 11 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 12', 'Task 12 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 13', 'Task 13 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 14', 'Task 14 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 15', 'Task 15 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 16', 'Task 16 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 17', 'Task 17 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 18', 'Task 18 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 19', 'Task 19 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'tasks'
      LIMIT
        1
    ), 'Task 20', 'Task 20 Desc', 1, '{}', CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;