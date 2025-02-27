BEGIN;

INSERT INTO
  products (
    unit_id,
    branch_id,
    code,
    factory_code,
    name,
    sku,
    barcode,
    specification,
    description,
    remark,
    price_sell,
    price_buy,
    margin,
    status,
    expired_at,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    1,
    'C0001',
    'FC0001',
    'DVR 8CH + HDD 4TB',
    'SK0001',
    'BC0001',
    NULL,
    NULL,
    NULL,
    4000000,
    3321000,
    16.975,
    1,
    NULL,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    1,
    '0002',
    '0002',
    'PC Server A',
    '0002',
    '0002',
    'PC Server A Specification',
    'PC Server A Description',
    'PC Server A Remark',
    10000000,
    9000000,
    10,
    1,
    NULL,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    1,
    '0003',
    '0003',
    'Product C',
    '0003',
    '0003',
    'Product C Specification',
    'Product C Description',
    'Product C Remark',
    3000,
    1200,
    60,
    1,
    NULL,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;