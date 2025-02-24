BEGIN;

INSERT INTO
  customers (
    customer_type_id,
    agent_id,
    code,
    name,
    address,
    phone,
    email,
    pic,
    status,
    created_by_id,
    created_at
  )
VALUES
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        groups.name = 'customer_types'
        AND mix_values.name = 'Buyer'
    ),
    NULL,
    'B1',
    'Buyer 1',
    'Buyer 1 Address',
    '4444444444',
    'buy@gmail.com',
    'Buyer 1 PIC',
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        groups.name = 'customer_types'
        AND mix_values.name = 'Supplier'
    ),
    NULL,
    'S1',
    'Seller 1',
    'Seller 1 Address',
    '5555555555',
    'sell@gmail.com',
    'Seller 1 PIC',
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        groups.name = 'customer_types'
        AND mix_values.name = 'Agent'
    ),
    NULL,
    'A1',
    'Agent 1',
    'Agent 1 Address',
    '6666666666',
    'agent@gmail.com',
    'Agent 1 PIC',
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        groups.name = 'customer_types'
        AND mix_values.name = 'Subcon'
    ),
    NULL,
    'SB1',
    'Subcon 1',
    'Subcon 1 Address',
    '7777777777',
    'subcon@gmail.com',
    'Subcon 1 PIC',
    1,
    1,
    CURRENT_TIMESTAMP
  ),
  (
    (
      SELECT
        mix_values.id
      FROM
        mix_values
        JOIN groups ON mix_values.group_id = groups.id
      WHERE
        groups.name = 'customer_types'
        AND mix_values.name = 'Buyer'
    ),
    3,
    'B2',
    'Buyer 2',
    'Buyer 2 Address',
    '34444444444',
    'buy2@gmail.com',
    'Buyer 2 PIC',
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;