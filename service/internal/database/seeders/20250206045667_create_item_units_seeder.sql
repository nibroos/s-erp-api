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
  );

COMMIT;