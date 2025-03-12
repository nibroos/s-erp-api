BEGIN;

INSERT INTO
  mix_values (
    group_id,
    name,
    description,
    status,
    options_json,
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
        name = 'warehouses'
      LIMIT
        1
    ), 
    'GUDANG UTAMA', 
    'Gudang Utama', 
    1, 
    '{"code": "WH001"}',
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
        name = 'warehouses'
      LIMIT
        1
    ), 
    'GUDANG TRANSIT', 
    'Gudang Transit', 
    1,
    '{"code": "WH002"}',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;