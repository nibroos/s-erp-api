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
    'superadmin',
    'Super Admin Role',
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
    'manager',
    'Manager Role',
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
    'marketing',
    'Marketing Role',
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
    'sales',
    'Sales Role',
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
    'purchase',
    'Purchase Role',
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
    'inventory',
    'Inventory Role',
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
    'production',
    'Production Role',
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
    'exim',
    'Exim Role',
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
    'accounting',
    'Accounting Role',
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
    'beacukai',
    'Beacukai Role',
    1,
    '{}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;