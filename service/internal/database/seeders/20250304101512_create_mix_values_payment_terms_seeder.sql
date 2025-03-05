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
        name = 'payment_terms'
      LIMIT
        1
    ), 
    'NET 30', 
    'Payment due within 30 days', 
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
        name = 'payment_terms'
      LIMIT
        1
    ), 
    'NET 45', 
    'Payment due within 45 days', 
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
        name = 'payment_terms'
      LIMIT
        1
    ), 
    'NET 60', 
    'Payment due within 60 days', 
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
        name = 'payment_terms'
      LIMIT
        1
    ), 
    'COD', 
    'Cash On Delivery', 
    1, 
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;
