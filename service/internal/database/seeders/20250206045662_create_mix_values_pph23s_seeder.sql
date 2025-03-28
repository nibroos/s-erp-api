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
        name = 'pph23s'
    ),
    'PPH23',
    'PPH23',
    2,
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
        name = 'pph23s'
    ),
    'PPH25',
    'PPH25',
    0.75,
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
        name = 'pph23s'
    ),
    'PPH26',
    'PPH26',
    20,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;