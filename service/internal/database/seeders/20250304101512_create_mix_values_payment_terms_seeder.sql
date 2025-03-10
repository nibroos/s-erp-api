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
    'WEEKLY', 
    'Payment due within 7 days', 
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
    'MONTHLY', 
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
    'BI-WEEKLY', 
    'Payment due within 14 days', 
    1, 
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;
