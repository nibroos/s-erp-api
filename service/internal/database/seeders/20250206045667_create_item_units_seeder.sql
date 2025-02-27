BEGIN;

INSERT INTO
  item_units (
    unit_id,
    ms_item_id,
    conversion,
    price_sell,
    price_buy,
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
    3571000,
    3700000,
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
    2100000,
    2300000,
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
    950000,
    1150000,
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
    300000,
    350000,
    1,
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;