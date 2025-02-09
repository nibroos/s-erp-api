BEGIN;

INSERT INTO
  mix_values (
    group_id,
    name,
    description,
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
    ), 'IDR', 'Indonesian Rupiah', 1, CURRENT_TIMESTAMP,
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
    ), 'USD', 'United States Dollar', 1, CURRENT_TIMESTAMP,
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
    ), 'EUR', 'Euro', 1, CURRENT_TIMESTAMP,
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
    ), 'SGD', 'Singapore Dollar', 1, CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;