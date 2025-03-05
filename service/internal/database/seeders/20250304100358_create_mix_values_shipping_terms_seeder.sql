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
        name = 'shipping_terms'
      LIMIT
        1
    ), 
    'FOB', 
    'Free On Board', 
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 
    'CIF', 
    'Cost Insurance and Freight', 
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
        name = 'shipping_terms'
      LIMIT
        1
    ), 
    'CFR', 
    'Cost and Freight', 
    1, 
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;