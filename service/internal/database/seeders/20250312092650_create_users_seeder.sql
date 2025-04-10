BEGIN;

INSERT INTO
  users (
    branch_id,
    username,
    email,
    name,
    password,
    address,
    created_at,
    updated_at
  )
VALUES
  (
    NULL,
    'cs',
    'cs@yubipro.com',
    'Customer Service User',
    crypt('csyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  -- technician
  (
    NULL,
    'tech',
    'tech@yubipro.com',
    'Technician User',
    crypt('techyubi', gen_salt('bf')),
    'YUBIPRO',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

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
        username = 'cs'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'customer_service'
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
        username = 'tech'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'technician'
    ),
    1,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;