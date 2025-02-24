BEGIN;

INSERT INTO
  vat_histories (
    vat_id,
    num,
    divider,
    multiplier,
    changed_at,
    status,
    remark,
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
        mix_values.name = '11%'
        AND groups.name = 'vats'
    ),
    11,
    1,
    0.1,
    CURRENT_TIMESTAMP,
    1,
    'Initial VAT',
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
        mix_values.name = '12%'
        AND groups.name = 'vats'
    ),
    12,
    1.2,
    0.1,
    CURRENT_TIMESTAMP,
    1,
    'Initial VAT',
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
        mix_values.name = '12%'
        AND groups.name = 'vats'
    ),
    12,
    1,
    0.1,
    CURRENT_TIMESTAMP,
    1,
    'VAT Hist 2 12%',
    1,
    CURRENT_TIMESTAMP
  );

COMMIT;