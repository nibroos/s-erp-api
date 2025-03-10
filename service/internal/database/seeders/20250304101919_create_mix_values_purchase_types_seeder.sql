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
        name = 'purchase_types'
      LIMIT
        1
    ), 
    'Production', 
    'Production materials', 
    1, 
    '{"code": "0x9d"}',
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
        name = 'purchase_types'
      LIMIT
        1
    ), 
    'Office', 
    'General office supplies', 
    1,
    '{"code": "0x92"}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;
