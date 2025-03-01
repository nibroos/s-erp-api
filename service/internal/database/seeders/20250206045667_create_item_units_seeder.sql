BEGIN;

INSERT INTO
  item_units (
    unit_id,
    product_id,
    conversion,
    price_sell,
    price_buy,
    margin,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    1,
    1,
    2000,
    800,
    60,
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
        groups.name = 'units'
        AND mix_values.name = 'BOX'
    ),
    1,
    0.002,
    1000000,
    400000,
    60,
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
        groups.name = 'units'
        AND mix_values.name = 'PACK'
    ),
    1,
    0.01,
    200000,
    80000,
    60,
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
        groups.name = 'units'
        AND mix_values.name = 'GRAM'
    ),
    2,
    1,
    400,
    200,
    100,
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
        groups.name = 'units'
        AND mix_values.name = 'KG'
    ),
    2,
    1000,
    400000,
    200000,
    100,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    3,
    1,
    5000,
    4500,
    10,
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
        groups.name = 'units'
        AND mix_values.name = 'CM'
    ),
    5,
    1,
    150,
    120,
    20,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    6,
    1,
    1600000,
    1520000,
    5,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    7,
    1,
    1100000,
    1000000,
    5,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    8,
    1,
    3700000,
    3571000,
    3.5,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    9,
    1,
    2300000,
    2100000,
    8,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    10,
    1,
    1150000,
    950000,
    17,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    11,
    1,
    350000,
    300000,
    14,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    12,
    1,
    4000000,
    3321000,
    16.975,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    13,
    1,
    10000000,
    9000000,
    10,
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
        groups.name = 'units'
        AND mix_values.name = 'PIECE'
    ),
    13,
    1,
    3000,
    1200,
    60,
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;