BEGIN;

INSERT INTO
  mix_values (
    group_id,
    parent_id,
    name,
    description,
    status,
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Packaging',
    'Packaging',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Material',
    'Material',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Barang Jadi',
    'Barang Jadi',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Barang Setengah Jadi',
    'Barang Setengah Jadi',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Bahan Penolong',
    'Bahan Penolong',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Office'
    ),
    'Peralatan Kantor',
    'Peralatan Kantor',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Fiber',
    'Fiber',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Mesin',
    'Mesin',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Peralatan Pabrik',
    'Peralatan Pabrik',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Scrap',
    'Scrap',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Production'
    ),
    'Reject',
    'Reject',
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
        name = 'item_sub_groups'
    ),
    (
      SELECT
        id
      FROM
        mix_values
      WHERE
        name = 'Service'
    ),
    'Maintenance',
    'Maintenance',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
  );

COMMIT;