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
        name = 'vats'
    ),
    '11%',
    '11% VAT',
    11,
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
        name = 'vats'
    ),
    '12%',
    '12% VAT',
    12,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;