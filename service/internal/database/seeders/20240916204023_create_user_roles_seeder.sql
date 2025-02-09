-- Insert user-role relationships into pools table
-- marketing, sales, purchasing, inventory, production, exim, accounting, beacukai
INSERT INTO
  pools (
    group1_id,
    group2_id,
    mv1_id,
    mv2_id,
    created_by_id,
    updated_by_id,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'admin'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'superadmin'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'manager'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'manager'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'marketing'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'marketing'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'sales'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'sales'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'purchasing'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'purchasing'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'inventory'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'inventory'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'production'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'production'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'exim'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'exim'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'accounting'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'accounting'
    ),
    1,
    1,
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
        name = 'users'
    ),
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'roles'
    ),
    (
      SELECT
        id
      FROM
        users
      WHERE
        username = 'beacukai'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'beacukai'
    ),
    1,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;