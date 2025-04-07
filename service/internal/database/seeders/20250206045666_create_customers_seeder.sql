BEGIN;

INSERT INTO
  customers (
    customer_type_id,
    agent_id,
    shortname,
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
    'Buy 1',
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
    'Sel 1',
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
    'Ag 1',
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
    'Sub 1',
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
    'Buy 2',
    'B2',
    'Buyer 2',
    'Buyer 2 Address',
    '34444444444',
    'buy2@gmail.com',
    'Buyer 2 PIC',
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
    'Yubi',
    'YBT1',
    'PT. Yubi Technology',
    'Gading Bukit Indah,
      Jl.Raya Gading Kirana Blok.G.5,
      RT.18 / RW.8,
      West Kelapa Gading,
      Kelapa Gading,
      North Jakarta City,
      Jakarta 14240',
    '111111',
    'pt.yubitechnology@gmail.com',
    'Yubi Technology PIC',
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;