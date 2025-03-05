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
    'LOCAL', 
    'Local Purchase', 
    1, 
    '{"code": "LCL"}',
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
    'IMPORT', 
    'Import Purchase', 
    1,
    '{"code": "IMP"}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;
