BEGIN;

-- marketing, sales, purchase, inventory, production, exim, accounting, beacukai
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
        name = 'roles'
    ),
    'technician',
    'Technician',
    1,
    '{}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    'customer_service',
    'Customer Service',
    1,
    '{}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    'developer',
    'Developer',
    1,
    '{}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;