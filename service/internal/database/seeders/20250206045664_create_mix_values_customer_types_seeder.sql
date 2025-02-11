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
        name = 'customer_types'
    ),
    'Buyer',
    'Buyer',
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
        name = 'customer_types'
    ),
    'Supplier',
    'Supplier',
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
        name = 'customer_types'
    ),
    'Agent',
    'Agent',
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
        name = 'customer_types'
    ),
    'Subcon',
    'Subcon',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;