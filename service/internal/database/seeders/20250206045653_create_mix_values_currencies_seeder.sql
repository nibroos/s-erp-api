BEGIN;

INSERT INTO
  mix_values (
    group_id,
    name,
    description,
    num,
    status,
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
        name = 'currencies'
      LIMIT
        1
    ), 'IDR', 'Indonesian Rupiah', 1, 1, CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'currencies'
      LIMIT
        1
    ), 'USD', 'United States Dollar', 16580, 1, CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'currencies'
      LIMIT
        1
    ), 'EUR', 'Euro', 17157.4, 1, CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'currencies'
      LIMIT
        1
    ), 'SGD', 'Singapore Dollar', 12262.3, 1, CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;