BEGIN;

INSERT INTO
  mix_values (
    group_id,
    name,
    description,
    status,
    created_at,
    updated_at
  ) -- buyer, supplier, agent, subcon
VALUES
  (
    (
      SELECT
        id
      FROM
        groups
      WHERE
        name = 'order_types'
    ),
    'Order',
    'Order',
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
        name = 'order_types'
    ),
    'Sampling',
    'Sampling',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;