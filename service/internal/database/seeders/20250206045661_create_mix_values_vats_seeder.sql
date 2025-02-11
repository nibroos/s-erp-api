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
    'VAT',
    'VAT Description',
    1,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;